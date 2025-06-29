// Package main provides a CLI application to process Apache log files.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/rbscholtus/go-webalizer/internal/charts"
	"github.com/rbscholtus/go-webalizer/internal/parser"
	"github.com/urfave/cli/v3"
)

func processFile(fileName string) error {
	// process log file
	stats, err := parser.ProcessLog(fileName)
	if err != nil {
		return err
	}

	// Monthy aggregates
	aggr := stats.AggregatesByMonth()

	// Render and save charts
	page := components.NewPage()
	page.AddCharts(charts.MonthlyBarCharts(aggr))

	f, err := os.Create("index.html")
	if err != nil {
		return err
	}
	page.Render(f)

	return nil
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
			return processFile(fileName)
		},
	}

	// Run the CLI command
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
