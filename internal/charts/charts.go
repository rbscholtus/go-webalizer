package charts

import (
	"sort"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/rbscholtus/go-webalizer/internal/logstats"
)

// Keys returns the keys of a map in sorted order
func SortedKeys[V any](m *map[string]V) []string {
	keys := make([]string, 0, len(*m))
	for key := range *m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func MonthlyBarCharts(aggr *map[string]*logstats.MonthData) (*charts.Bar, *charts.Bar, *charts.Bar) {
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

	numMonths := len(*aggr)
	months := make([]string, 0, numMonths)

	// calculate series data
	hits := make([]opts.BarData, 0, numMonths)
	files := make([]opts.BarData, 0, numMonths)
	pages := make([]opts.BarData, 0, numMonths)
	bytes := make([]opts.BarData, 0, numMonths)
	visits := make([]opts.BarData, 0, numMonths)
	sites := make([]opts.BarData, 0, numMonths)

	keys := SortedKeys(aggr)
	for _, key := range keys {
		data := (*aggr)[key]
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

	return hfpBar, bBar, vsBar
}
