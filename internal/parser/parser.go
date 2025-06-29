package parser

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/rbscholtus/go-webalizer/internal/logstats"
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

// processLog parses the log file line-by-line and accumulates stats.
func ProcessLog(fileName string) (*logstats.LogStats, error) {
	// Open the access log file
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	lineNr := 0
	stats := logstats.NewLogStats()
	line := LogEntry{}

	fileExtRE := regexp.MustCompile(fileExts)
	// var dumper = godump.Dumper{Theme: godump.DefaultTheme}

	// Scan the log line-by-line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// scan and parse a line
		lineNr++
		ok, err := line.Extract(scanner.Bytes())
		if !ok {
			fmt.Fprintln(os.Stderr, "Invalid line", lineNr, ":", err)
			// dumper.Fprintln(os.Stderr, line)
			continue
		}

		// dumper.Fprintln(os.Stderr, line)
		// break

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

		// VISITS: Determine if this is a new "visit" based on timeout
		if line.Timestamp.Sub(stats.LastVisit[line.IP]) > visitTimeout {
			if _, ok := stats.Visits[date]; !ok {
				stats.Visits[date] = make(map[string]uint64)
			}
			stats.Visits[date][line.IP]++
			incVisits = true
		}

		// Track first and last hit time
		if _, ok := stats.FirstVisit[line.IP]; !ok {
			stats.FirstVisit[line.IP] = line.Timestamp
		}
		stats.LastVisit[line.IP] = line.Timestamp

		// SITES: Count hits by IP
		if _, ok := stats.Sites[date]; !ok {
			stats.Sites[date] = make(map[string]uint64)
		}
		stats.Sites[date][line.IP]++

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
		stats.UpdateIPStats(date, line.IP, line.Size, incVisits)

		// USERAGENTS: Reports hits, bytes, and visits by User-Agent
		stats.UpdateUserAgentStats(date, line.UserAgent, line.Size, incVisits)

		// URLPaths: Report hits and bytes by URLPath and Method
		stats.UpdateURLStats(date, line.URLPath, line.Method, line.Size)

		// REFERRERS: Reports hits and bytes by Referrer
		stats.UpdateReferrerStats(date, line.Referrer, line.Size)
	}

	// Report any errors from scanning
	if err := scanner.Err(); err != nil {
		msg := fmt.Errorf("error reading file: %v", err)
		return nil, msg
	}

	return stats, nil
}
