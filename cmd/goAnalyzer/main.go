// Package main provides a CLI application to process Apache log files.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/urfave/cli/v3"
)

// Stats holds aggregated metrics parsed from web server log files.
type Stats struct {
	// Hits is the total number of requests (including all file types).
	Hits int

	// Files is the count of successful file responses (e.g., images, scripts, documents).
	Files int

	// Pages is the number of page views (based on HTML extensions or predefined URLs).
	Pages int

	// SiteNames maps each remote address (IP or host) to the number of requests it made.
	SiteNames map[string]int

	// LastVisit stores the most recent timestamp seen for each site.
	LastVisit map[string]time.Time

	// Visits is the number of unique visits based on IP and timeout threshold.
	Visits int

	// Bytes is the total number of bytes sent by the server.
	Bytes int

	// HitsByMethod counts requests grouped by HTTP method (GET, POST, etc.).
	HitsByMethod map[string]int

	// HitsByProtocol counts requests grouped by HTTP protocol version.
	HitsByProtocol map[string]int

	// HitsByRespCode counts requests grouped by HTTP response status code.
	HitsByRespCode map[string]int
}

// NewStats initializes and returns a new Stats object.
func NewStats() *Stats {
	stats := Stats{}
	stats.SiteNames = make(map[string]int)
	stats.LastVisit = make(map[string]time.Time)
	stats.HitsByMethod = make(map[string]int)
	stats.HitsByProtocol = make(map[string]int)
	stats.HitsByRespCode = make(map[string]int)
	return &stats
}

// String prints a human-readable summary of stats.
func (s *Stats) String() string {
	output := fmt.Sprintf("Stats:\n"+
		"  Hits:   %d\n"+
		"  Files:  %d\n"+
		"  Pages:  %d\n"+
		"  Sites:  %d\n"+
		"  Visits: %d\n"+
		"  Bytes:  %d bytes\n",
		s.Hits, s.Files, s.Pages, len(s.SiteNames), s.Visits, s.Bytes)

	output += fmt.Sprintln("  Hits by HTTP method")
	for _, method := range GetSortedKeys(&s.HitsByMethod) {
		output += fmt.Sprintf("    %-7s: %d\n", method, s.HitsByMethod[method])
	}

	output += fmt.Sprintln("  Hits by HTTP protocol version")
	for _, protocol := range GetSortedKeys(&s.HitsByProtocol) {
		output += fmt.Sprintf("    %-8s: %d\n", protocol, s.HitsByProtocol[protocol])
	}

	output += fmt.Sprintln("  Hits by HTTP response code")
	for _, retCode := range GetSortedKeys(&s.HitsByRespCode) {
		output += fmt.Sprintf("    %s: %d\n", retCode, s.HitsByRespCode[retCode])
	}

	return output
}

// GetSortedKeys returns sorted keys of a map for ordered output.
func GetSortedKeys(m *map[string]int) []string {
	keys := make([]string, 0, len(*m))
	for k := range *m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// processLog parses the log file line-by-line and accumulates stats.
func processLog(fileName string) {
	// Open the access log file.
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	lineNr := 0
	stats := NewStats()

	// Regular expression to parse each log line (Common Log Format + user-agent)
	//                         1     2     3       4        5        6        7                  8     9           10      11
	re := regexp.MustCompile(`^(\S+) (\S+) (\S+) $begin:math:display$(.*?)$end:math:display$ "([A-Z]+) (.*?)(?: (HTTP\/[0-9.]+))?" (\d+) (-|\d+)(?: "(.*?)" "(.*?)")?$`)

	// Regex to detect known HTML-like file extensions (used for "page" classification)
	re2 := regexp.MustCompile(`(?i)\.(htm|html|...|hbs)`)

	// Additional patterns considered "page views" even if no extension
	PageURLRegExes := []string{"^/$", "^/blog", "^/articles", "^/projects"}
	var compiledRegexes []*regexp.Regexp
	for _, regexStr := range PageURLRegExes {
		compiledRegex, err := regexp.Compile(regexStr)
		if err != nil {
			fmt.Println("Invalid regex:", regexStr)
			continue
		}
		compiledRegexes = append(compiledRegexes, compiledRegex)
	}

	// Time layout used to parse log timestamps
	layout := "02/Jan/2006:15:04:05 -0700"
	visitTimeout := 600 * time.Second // 10-minute session timeout for a "new visit"

	// Scan the log line-by-line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNr++
		match := re.FindStringSubmatch(scanner.Text())
		if match == nil {
			fmt.Println("Invalid line", lineNr)
			continue
		}

		// Every successfully parsed line is a hit
		stats.Hits++

		// Increment files for successful responses (HTTP 200)
		if match[8] == "200" {
			stats.Files++
		}

		// Classify as a "page" by extension
		if re2.FindStringIndex(match[6]) != nil {
			stats.Pages++
		} else {
			// Or match predefined page-like URL patterns
			for _, re := range compiledRegexes {
				if re.FindStringIndex(match[6]) != nil {
					stats.Pages++
				}
			}
		}

		// Count visits by IP and track last access time
		ip := match[1]
		stats.SiteNames[ip]++

		// Parse timestamp from log
		t, err := time.Parse(layout, match[4])
		if err != nil {
			fmt.Println("Error parsing time:", err)
			return
		}

		// Determine if this is a new "visit" based on timeout
		if t.Sub(stats.LastVisit[ip]) > visitTimeout {
			stats.Visits++
		}
		stats.LastVisit[ip] = t

		// Track total bytes sent (if numeric)
		bytes, err := strconv.Atoi(match[9])
		if err == nil {
			stats.Bytes += bytes
		}

		// Track method, protocol, and status code
		stats.HitsByMethod[match[5]]++
		if match[7] == "" {
			stats.HitsByProtocol["Unknown"]++
		} else {
			stats.HitsByProtocol[match[7]]++
		}
		stats.HitsByRespCode[match[8]]++
	}

	// Report any errors from scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	// Print final statistics
	fmt.Println(stats)
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
