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

type Stats struct {
	hits           int                  // Any request made to the server is a 'hit'
	files          int                  // Responses sent back to the requesting client (HTML page, image...)
	pages          int                  // Any HTML document—or anything that generates an HTML document—would
	siteNames      map[string]int       // Each request made to the server comes from a unique 'site'
	lastVisit      map[string]time.Time // Last hit' timestamp per site
	visits         int                  // Number of new or timed out IP addresses
	bytes          int                  // Amount of data that was sent out by the server
	hitsByMethod   map[string]int       // Hits by HTTP method
	hitsByProtocol map[string]int       // Hits by HTTP protocol
	hitsByRespCode map[string]int       // Hits by HTTP response code
}

func NewStats() *Stats {
	stats := Stats{}
	stats.siteNames = make(map[string]int)
	stats.lastVisit = make(map[string]time.Time)
	stats.hitsByMethod = make(map[string]int)
	stats.hitsByProtocol = make(map[string]int)
	stats.hitsByRespCode = make(map[string]int)
	return &stats
}

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

func GetSortedKeys(m *map[string]int) []string {
	keys := make([]string, 0, len(*m))
	for k := range *m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func processLog(fileName string) {
	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	lineNr := 0
	stats := NewStats()

	// How to recognise a Page
	//                         1     2     3       4        5        6        7                  8     9           10      11
	re := regexp.MustCompile(`^(\S+) (\S+) (\S+) \[(.*?)\] "([A-Z]+) (.*?)(?: (HTTP\/[0-9.]+))?" (\d+) (-|\d+)(?: "(.*?)" "(.*?)")?$`)
	re2 := regexp.MustCompile(`(?i)\.(htm|html|htmlx|dhtml|phtml|php3|php4|php|asp|aspx|cfm|cfml|jsp|jspx|shtml|stm|cgi|pl|py|ejs|erb|haml|handlebars|mustache|twig|jade|pug|dust|liquid|md|markdown|html\.erb|rst|adoc|hbs)`)

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

	layout := "02/Jan/2006:15:04:05 -0700"
	visitTimeout := 600 * time.Second

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNr++

		match := re.FindStringSubmatch(scanner.Text())
		if match == nil {
			fmt.Println("Invalid line", lineNr)
			continue
		}

		// any valid log entry is a hit
		stats.hits++

		// "Found" increases files
		if match[8] == "200" {
			stats.files++
		}

		// any of the predefined file extensions is a page
		if re2.FindStringIndex(match[6]) != nil {
			stats.pages++
		} else {
			// any of the predefined Page URLs is also a page
			for _, re := range compiledRegexes {
				if re.FindStringIndex(match[6]) != nil {
					stats.pages++
				}
			}
		}

		// count visits
		stats.siteNames[match[1]]++

		t, err := time.Parse(layout, match[4])
		if err != nil {
			fmt.Println("Error parsing time:", err)
			return
		}
		if t.Sub(stats.lastVisit[match[1]]) > visitTimeout {
			stats.visits++
		}
		stats.lastVisit[match[1]] = t

		// the number of bytes sent back
		bytes, err := strconv.Atoi(match[9])
		if err == nil {
			stats.bytes += bytes
		}

		// count hits by HTTP method, protocol, response code
		stats.hitsByMethod[match[5]]++
		if match[7] == "" {
			stats.hitsByProtocol["Unknown"]++
		} else {
			stats.hitsByProtocol[match[7]]++
		}
		stats.hitsByRespCode[match[8]]++
	}

	// Check for errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	// print
	fmt.Println(stats)
}

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

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

}
