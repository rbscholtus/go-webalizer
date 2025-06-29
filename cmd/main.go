// Package main provides a CLI application to process Apache log files.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/rbscholtus/go-webalizer/internal/parser"
	"github.com/urfave/cli/v3"
)

// Keys returns the keys of a map in sorted order
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type MonthData struct {
	Month  string
	Hits   uint64
	Files  uint64
	Pages  uint64
	Sites  uint64
	Visits uint64
	Bytes  uint64
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
			stats, err := parser.ProcessLog(fileName)
			if err != nil {
				return nil
			}

			// Monthy aggregates
			aggr := make(map[string]*MonthData)
			for dateStr, hits := range stats.Hits {
				files := stats.Files[dateStr]
				pages := stats.Pages[dateStr]
				sites := uint64(len(stats.Sites[dateStr]))
				visits := uint64(0)
				for _, count := range stats.Visits[dateStr] {
					visits += count
				}
				bytes := stats.Bytes[dateStr]

				monthStr := dateStr[:7]
				value, ok := aggr[monthStr]
				if !ok {
					date, _ := time.Parse("2006-01", monthStr)
					value = &MonthData{date.Format("Jan"), hits, files, pages, sites, visits, bytes}
					aggr[monthStr] = value
				} else {
					value.Hits += hits
					value.Files += files
					value.Pages += pages
					value.Sites += sites
					value.Visits += visits
					value.Bytes += bytes
				}
			}

			// Render and save charts
			page := components.NewPage()
			page.AddCharts(monthlyBarCharts(aggr))

			f, err := os.Create("index.html")
			if err == nil {
				page.Render(f)
			}

			return nil
		},
	}

	// Run the CLI command
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func monthlyBarCharts(aggr map[string]*MonthData) (*charts.Bar, *charts.Bar, *charts.Bar) {
	// common options
	xAxisOpts := opts.XAxis{
		SplitLine: &opts.SplitLine{
			Show: opts.Bool(true),
		},
	}
	yAxisOpts := opts.YAxis{
		SplitLine: &opts.SplitLine{
			Show: opts.Bool(true),
		},
	}
	gapOpt := charts.WithBarChartOpts(opts.BarChart{
		BarGap: "-75%",
	})
	styleOpt := charts.WithItemStyleOpts(opts.ItemStyle{
		BorderWidth: 1,
		BorderColor: "black",
	})

	months := make([]string, 0, len(aggr))

	// calculate series data
	hits := make([]opts.BarData, 0, len(aggr))
	files := make([]opts.BarData, 0, len(aggr))
	pages := make([]opts.BarData, 0, len(aggr))
	sites := make([]opts.BarData, 0, len(aggr))
	visits := make([]opts.BarData, 0, len(aggr))
	bytes := make([]opts.BarData, 0, len(aggr))

	keys := SortedKeys(aggr)
	for _, key := range keys {
		data := aggr[key]
		months = append(months, data.Month)

		hits = append(hits, opts.BarData{Value: data.Hits})
		files = append(files, opts.BarData{Value: data.Files})
		pages = append(pages, opts.BarData{Value: data.Pages})
		sites = append(sites, opts.BarData{Value: data.Sites})
		visits = append(visits, opts.BarData{Value: data.Visits})
		bytes = append(bytes, opts.BarData{Value: data.Bytes})
	}

	// create 3 Bar charts
	hfpBar := charts.NewBar()
	hfpBar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Usage summary"}), // Subtitle: "This is the subtitle."
		charts.WithColorsOpts(opts.Colors{"#00805c", "#0040ff", "#00e0ff"}),
		charts.WithXAxisOpts(xAxisOpts),
		charts.WithYAxisOpts(yAxisOpts),
	)
	hfpBar.SetXAxis(months).
		AddSeries("Hits", hits).
		AddSeries("Files", files).
		AddSeries("Pages", pages)
	hfpBar.SetSeriesOptions(gapOpt, styleOpt)

	vsBar := charts.NewBar()
	vsBar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Usage summary"}), // Subtitle: "This is the subtitle."
		charts.WithColorsOpts(opts.Colors{"#ffff00", "#ff8000"}),
		charts.WithXAxisOpts(xAxisOpts),
		charts.WithYAxisOpts(yAxisOpts),
	)
	vsBar.SetXAxis(months).
		AddSeries("Visits", visits).
		AddSeries("Sites", sites)
	vsBar.SetSeriesOptions(gapOpt, styleOpt)

	bBar := charts.NewBar()
	bBar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Usage summary"}), // Subtitle: "This is the subtitle."
		charts.WithColorsOpts(opts.Colors{"#ff0000"}),
		charts.WithXAxisOpts(xAxisOpts),
		charts.WithYAxisOpts(yAxisOpts),
	)
	bBar.SetXAxis(months).
		AddSeries("Bytes", bytes)
	bBar.SetSeriesOptions(gapOpt, styleOpt)

	return hfpBar, vsBar, bBar
}
