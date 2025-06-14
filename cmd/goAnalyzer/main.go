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

// Stats holds various aggregated metrics from parsed log entries.
type Stats struct {
	hits           int                  // Total HTTP requests (lines parsed successfully)
	files          int                  // Number of successful responses (e.g., HTTP 200)
	pages          int                  // Number of HTML/doc-like page accesses
	siteNames      map[string]int       // Requests per IP/site
	lastVisit      map[string]time.Time // Last request time per IP/site
	visits         int                  // Distinct visits (timeout-based session)
	bytes          int                  // Total bytes sent by the server
	hitsByMethod   map[string]int       // Count by HTTP method (GET, POST, etc.)
	hitsByProtocol map[string]int       // Count by HTTP protocol version
	hitsByRespCode map[string]int       // Count by HTTP response code
}

// NewStats initializes and returns a new Stats object.
func NewStats() *Stats {
	stats := Stats{}
	stats.siteNames = make(map[string]int)
	stats.lastVisit = make(map[string]time.Time)
	stats.hitsByMethod = make(map[string]int)
	stats.hitsByProtocol = make(map[string]int)
	stats.hitsByRespCode = make(map[string]int)
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
		s.hits, s.files, s.pages, len(s.siteNames), s.visits, s.bytes)

	output += fmt.Sprintln("  Hits by HTTP method")
	for _, method := range GetSortedKeys(&s.hitsByMethod) {
		output += fmt.Sprintf("    %-7s: %d\n", method, s.hitsByMethod[method])
	}

	output += fmt.Sprintln("  Hits by HTTP protocol")
	for _, protocol := range GetSortedKeys(&s.hitsByProtocol) {
		output += fmt.Sprintf("    %-8s: %d\n", protocol, s.hitsByProtocol[protocol])
	}

	output += fmt.Sprintln("  Hits by HTTP response code")
	for _, retCode := range GetSortedKeys(&s.hitsByRespCode) {
		output += fmt.Sprintf("    %s: %d\n", retCode, s.hitsByRespCode[retCode])
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
		stats.hits++

		// Increment files for successful responses (HTTP 200)
		if match[8] == "200" {
			stats.files++
		}

		// Classify as a "page" by extension
		if re2.FindStringIndex(match[6]) != nil {
			stats.pages++
		} else {
			// Or match predefined page-like URL patterns
			for _, re := range compiledRegexes {
				if re.FindStringIndex(match[6]) != nil {
					stats.pages++
				}
			}
		}

		// Count visits by IP and track last access time
		ip := match[1]
		stats.siteNames[ip]++

		// Parse timestamp from log
		t, err := time.Parse(layout, match[4])
		if err != nil {
			fmt.Println("Error parsing time:", err)
			return
		}

		// Determine if this is a new "visit" based on timeout
		if t.Sub(stats.lastVisit[ip]) > visitTimeout {
			stats.visits++
		}
		stats.lastVisit[ip] = t

		// Track total bytes sent (if numeric)
		bytes, err := strconv.Atoi(match[9])
		if err == nil {
			stats.bytes += bytes
		}

		// Track method, protocol, and status code
		stats.hitsByMethod[match[5]]++
		if match[7] == "" {
			stats.hitsByProtocol["Unknown"]++
		} else {
			stats.hitsByProtocol[match[7]]++
		}
		stats.hitsByRespCode[match[8]]++
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
