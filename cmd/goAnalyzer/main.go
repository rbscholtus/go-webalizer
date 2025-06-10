package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/urfave/cli/v2"
)

func processLog(fileName string) {
	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	var count int = 0

	/* var (
		host       string
		dash       string
		user       string
		timestamp  string
		request    string
		statusCode int
		bytes      int
		referrer   string
		useragent  string
	) */

	re := regexp.MustCompile(`^(\S+) (\S+) (\S+) \[(.*?)\] "(.*?)" (\d+) (\d+) "(.*?)" "(.*?)"$`)

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			match := re.FindStringSubmatch(line)

			if match == nil {
				fmt.Println("No match")
				return
			}

			count++

			fmt.Printf("Host: %s\n", match[1])
			fmt.Printf("Dash: %s\n", match[2])
			fmt.Printf("User: %s\n", match[3])
			fmt.Printf("Timestamp: %s\n", match[4])
			fmt.Printf("Request: %s\n", match[5])
			fmt.Printf("Status Code: %s\n", match[6])
			fmt.Printf("Bytes: %s\n", match[7])
			fmt.Printf("Referrer: %s\n", match[8])
			fmt.Printf("User-Agent: %s\n", match[9])

			return
		}
	}

	// Check for errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

}

func main() {
	app := &cli.App{
		Name:  "file-cli",
		Usage: "A simple CLI that takes a file name as an argument",
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return fmt.Errorf("please provide exactly one file name")
			}

			fileName := c.Args().Get(0)
			processLog(fileName)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
