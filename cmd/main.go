// Package main provides a CLI application to process Apache log files.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rbscholtus/go-webalizer/internal/parser"
	"github.com/urfave/cli/v3"
)

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
			parser.ProcessLog(fileName)
			return nil
		},
	}

	// Run the CLI command
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
