// Package main provides a CLI application to process Apache log files.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/yassinebenaid/godump"
)

// the datetime format of the log timestamp
const dateFormat = "02/Jan/2006:15:04:05 -0700"

// 10-minute session timeout for a "new visit"
const visitTimeout = 600 * time.Second

// extensions of files that resemble a "page"
const fileExts = `\.(htm|html|php|php3|php4|asp|aspx|jsp|js|py|shtml|xhtml|cgi|pl|rb|erb|ejs|phtml|dhtml|cfm|do|action|axd|ashx|asmx|svc|faces|jspx|xsp|md|markdown|liquid|mustache|hbs|wsdl|wadl|swagger)`

// unmarshalIP converts a IP/DNS string from a log entry.
func (p *LogEntry) unmarshalIP(value []byte) (string, error) {
	return string(value), nil
}

// unmarshalTimestamp parses a timestamp from a log entry.
func (p *LogEntry) unmarshalTimestamp(value []byte) (time.Time, error) {
	// Parse timestamp using the predefined date format
	return time.Parse(dateFormat, string(value))
}

// unmarshalMethod converts a Method string from a log entry.
func (p *LogEntry) unmarshalMethod(value []byte) (string, error) {
	return string(value), nil
}

// unmarshalURLPath unescapes a URL path from a log entry.
func (p *LogEntry) unmarshalURLPath(value []byte) (string, error) {
	unescapedPath, err := url.PathUnescape(string(value))
	if err != nil {
		return string(value), err
	}
	return unescapedPath, nil
}

// unmarshalSize parses the size of a response from a log entry.
func (p *LogEntry) unmarshalSize(value []byte) (uint64, error) {
	// Check for a dash (-) indicating an unknown or missing size
	if string(value) == "-" {
		return 0, nil
	}
	return strconv.ParseUint(string(value), 10, 64)
}

// unmarshalReferrer unescapes a referrer URL from a log entry.
func (p *LogEntry) unmarshalReferrer(value []byte) (string, error) {
	unescapedRef, err := url.PathUnescape(string(value))
	if err != nil {
		return string(value), err
	}
	return unescapedRef, nil
}

// unmarshalUserAgent converts a UserAgent string from a log entry.
func (p *LogEntry) unmarshalUserAgent(value []byte) (string, error) {
	return string(value), nil
}

// Stats holds aggregated metrics parsed from web server log files.
type HB struct {
	Hits  uint64
	Bytes uint64
}

func (h *HB) AddTraffic(bytes uint64) {
	h.Hits++
	h.Bytes += bytes
}

type HBV struct {
	Hits   uint64
	Bytes  uint64
	Visits uint64
}

func (h *HBV) AddTraffic(bytes uint64, isNewVisit bool) {
	h.Hits++
	h.Bytes += bytes
	if isNewVisit {
		h.Visits++
	}
}
func getOrCreateHBV(IPs map[string]map[string]*HBV, date string, ip string) *HBV {
	if IPs[date] == nil {
		IPs[date] = make(map[string]*HBV)
	}
	if _, ok := IPs[date][ip]; !ok {
		IPs[date][ip] = &HBV{}
	}
	return IPs[date][ip]
}

func updateHBVStats(data map[string]map[string]*HBV, date string, field string, bytes uint64, isNewVisit bool) {
	getOrCreateHBV(data, date, field).AddTraffic(bytes, isNewVisit)
}

func getOrCreateURLHB(URLs map[string]map[string]map[string]*HB, date string, ip string, method string) *HB {
	if URLs[date] == nil {
		URLs[date] = make(map[string]map[string]*HB)
	}
	if _, ok := URLs[date][ip]; !ok {
		URLs[date][ip] = make(map[string]*HB)
	}
	if _, ok := URLs[date][ip][method]; !ok {
		URLs[date][ip][method] = &HB{}
	}
	return URLs[date][ip][method]
}

func updateURLStats(URLs map[string]map[string]map[string]*HB, date string, ip string, method string, bytes uint64) {
	getOrCreateURLHB(URLs, date, ip, method).AddTraffic(bytes)
}

func getOrCreateRefHB(referrers map[string]map[string]*HB, date string, referrer string) *HB {
	if referrers[date] == nil {
		referrers[date] = make(map[string]*HB)
	}
	if _, ok := referrers[date][referrer]; !ok {
		referrers[date][referrer] = &HB{}
	}
	return referrers[date][referrer]
}

func updateReferrerStats(referrers map[string]map[string]*HB, date string, referrer string, bytes uint64) {
	getOrCreateRefHB(referrers, date, referrer).AddTraffic(bytes)
}

type LogStats struct {
	Hits       map[string]uint64 // Key: "YYYY-MM-DD HH", "YYYY-MM-DD", "YYYY-MM"
	Files      map[string]uint64
	Pages      map[string]uint64
	Bytes      map[string]uint64
	Visits     map[string]map[string]uint64         // Key: "YYYY-MM-DD HH", "YYYY-MM-DD", "YYYY-MM", value: map[IP]uint64
	Sites      map[string]map[string]uint64         // Key: "YYYY-MM-DD HH", "YYYY-MM-DD", "YYYY-MM", value: map[IP]uint64
	Methods    map[string]map[string]uint64         // Key: "YYYY-MM-DD", "YYYY-MM-DD", "YYYY-MM", value: map[Method]uint64
	RespCodes  map[string]map[uint16]uint64         // Key: "YYYY-MM-DD", "YYYY-MM-DD", "YYYY-MM", value: map[Response Code]uint64
	IPs        map[string]map[string]*HBV           // Key: "YYYY-MM-DD HH", value: map[IP]*HBV
	UserAgents map[string]map[string]*HBV           // Key: "YYYY-MM-DD HH", value: map[UserAgent]struct { Hits uint64; Visits uint64; Bytes uint64 }
	URLPaths   map[string]map[string]map[string]*HB // Key: "YYYY-MM-DD HH", value: map[URLPath], value: map[Method]struct { Hits uint64; Bytes uint64 }
	Referrers  map[string]map[string]*HB            // Key: "YYYY-MM-DD HH", value: map[Referrer]struct { Hits uint64; Bytes uint64 }
	lastVisit  map[string]time.Time
}

func NewLogStats() *LogStats {
	return &LogStats{
		Hits:       make(map[string]uint64),
		Files:      make(map[string]uint64),
		Pages:      make(map[string]uint64),
		Bytes:      make(map[string]uint64),
		Visits:     make(map[string]map[string]uint64),
		Sites:      make(map[string]map[string]uint64),
		Methods:    make(map[string]map[string]uint64),
		RespCodes:  make(map[string]map[uint16]uint64),
		IPs:        make(map[string]map[string]*HBV),
		UserAgents: make(map[string]map[string]*HBV),
		URLPaths:   make(map[string]map[string]map[string]*HB),
		Referrers:  make(map[string]map[string]*HB),
		lastVisit:  make(map[string]time.Time),
	}
}

// processLog parses the log file line-by-line and accumulates stats.
func processLog(fileName string) {
	// Open the access log file
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	lineNr := 0
	stats := NewLogStats()
	line := LogEntry{}

	fileExtRE := regexp.MustCompile(fileExts)

	// Scan the log line-by-line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// scan and parse a line
		lineNr++
		ok, err := line.Extract(scanner.Bytes())
		if !ok {
			fmt.Println("Invalid line", lineNr, ":", err)
			// godump.Dump(line)
			continue
		}

		// If Visits was incremented for this log line
		incVisits := false

		date := line.Timestamp.Format("2006-01-02")

		// HITS: Every successfully parsed line is a hit
		stats.Hits[date]++

		// FILES: Increment files for successful responses (HTTP 200)
		if line.RespCode == 200 {
			stats.Files[date]++
		}

		// PAGES: Classify as a "page" by extension
		if fileExtRE.FindStringIndex(line.URLPath) != nil {
			stats.Pages[date]++
		}
		// else {
		// 	// Or match predefined page-like URL patterns
		// 	for _, re := range compiledRegexes {
		// 		if re.FindStringIndex(match[6]) != nil {
		// 			stats.Pages++
		// 		}
		// 	}
		// }

		// BYTES: Track total bytes sent (if numeric)
		stats.Bytes[date] += line.Size

		// SITES: Count visits by IP and track last access time
		if _, ok := stats.Sites[date]; !ok {
			stats.Sites[date] = make(map[string]uint64)
		}
		stats.Sites[date][line.IP]++

		// VISITS: Determine if this is a new "visit" based on timeout
		if line.Timestamp.Sub(stats.lastVisit[line.IP]) > visitTimeout {
			if _, ok := stats.Visits[date]; !ok {
				stats.Visits[date] = make(map[string]uint64)
			}
			stats.Visits[date][line.IP]++
			incVisits = true
		}
		stats.lastVisit[line.IP] = line.Timestamp

		// METHOD: count hits by method
		if _, ok := stats.Methods[date]; !ok {
			stats.Methods[date] = make(map[string]uint64)
		}
		stats.Methods[date][line.Method]++

		// METHOD: count hits by response code
		if _, ok := stats.RespCodes[date]; !ok {
			stats.RespCodes[date] = make(map[uint16]uint64)
		}
		stats.RespCodes[date][line.RespCode]++

		// IPs: Reports hits, bytes, and visits by IP
		updateHBVStats(stats.IPs, date, line.IP, line.Size, incVisits)

		// USERAGENTS: Reports hits, bytes, and visits by User-Agent
		updateHBVStats(stats.UserAgents, date, line.UserAgent, line.Size, incVisits)

		// URLPaths: Report hits and bytes by URLPath and Method
		updateURLStats(stats.URLPaths, date, line.URLPath, line.Method, line.Size)

		// REFERRERS: Reports hits and bytes by Referrer
		updateReferrerStats(stats.Referrers, date, line.Referrer, line.Size)
	}

	// Report any errors from scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	// Print final statistics
	d := godump.Dumper{HidePrivateFields: true}
	d.Println(stats)
}

// main defines and runs the CLI using urfave/cli.
func main() {
	cmd := &cli.Command{
		Name:  "file-cli",
		Usage: "A simple CLI that takes a file name as an argument",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.NArg() != 1 {
				return fmt.Errorf("please provide exactly one file name")
			}
			fileName := cmd.Args().Get(0)
			processLog(fileName)
			return nil
		},
	}

	// Run the CLI command
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
