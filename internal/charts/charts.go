// Package charts provides functions for generating various charts based on web server log data.
package charts

import (
	"fmt"
	"maps"
	"slices"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/rbscholtus/go-webalizer/internal/http"
	"github.com/rbscholtus/go-webalizer/internal/logstats"
)

// MonthlyBarCharts generates three bar charts for monthly hits, files, pages, bytes, visits, and sites.
func MonthlyBarCharts(aggr map[string]*logstats.HFPBVSData) (*charts.Bar, *charts.Bar, *charts.Bar) {
	// Define common options for the charts.
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

	// Calculate series data for the charts.
	numMonths := len(aggr)
	months := make([]string, 0, numMonths)
	hits := make([]opts.BarData, 0, numMonths)
	files := make([]opts.BarData, 0, numMonths)
	pages := make([]opts.BarData, 0, numMonths)
	bytes := make([]opts.BarData, 0, numMonths)
	visits := make([]opts.BarData, 0, numMonths)
	sites := make([]opts.BarData, 0, numMonths)

	// Get the sorted keys of the aggregate map.
	keys := slices.Sorted(maps.Keys(aggr))
	for _, key := range keys {
		data := aggr[key]
		months = append(months, data.Category)

		hits = append(hits, opts.BarData{Value: data.Hits})
		files = append(files, opts.BarData{Value: data.Files})
		pages = append(pages, opts.BarData{Value: data.Pages})
		sites = append(sites, opts.BarData{Value: data.Sites})
		visits = append(visits, opts.BarData{Value: data.Visits})
		bytes = append(bytes, opts.BarData{Value: data.Bytes})
	}

	// Create three bar charts.
	hfpBar := charts.NewBar()
	hfpBar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Usage summary"}),
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
		charts.WithTitleOpts(opts.Title{Title: "Usage summary"}),
		charts.WithColorsOpts(opts.Colors{"#ff0000"}),
		charts.WithXAxisOpts(xAxisOpts),
		charts.WithYAxisOpts(yAxisOpts),
	)
	bBar.SetXAxis(months).
		AddSeries("Bytes", bytes)
	bBar.SetSeriesOptions(gapOpt, styleOpt)

	vsBar := charts.NewBar()
	vsBar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Usage summary"}),
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

// MethodPieChart generates a pie chart for HTTP method distribution.
func MethodPieChart(aggr map[string]uint64) *charts.Pie {
	pie := charts.NewPie()

	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Hits by HTTP Method",
		}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	// Calculate series data for the chart.
	items := make([]opts.PieData, 0, len(aggr))
	for meth, hits := range aggr {
		items = append(items, opts.PieData{Name: meth, Value: hits})
	}

	pie.AddSeries("Method", items).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show:      opts.Bool(true),
				Formatter: "{b} ({d}%)",
			}),
			charts.WithPieChartOpts(opts.PieChart{
				Radius: []string{"30%", "75%"},
			}),
		)

	return pie
}

// ResponsesPieChart generates a pie chart for HTTP response code distribution.
func ResponsesPieChart(aggr map[uint16]uint64) *charts.Pie {
	pie := charts.NewPie()

	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Hits by Response code",
		}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	// Calculate series data for the chart.
	items := make([]opts.PieData, 0, len(aggr))
	for code, hits := range aggr {
		key := fmt.Sprintf("%d - %s", code, http.HttpStatusCodes[code])
		items = append(items, opts.PieData{Name: key, Value: hits})
	}

	pie.AddSeries("Response Code", items).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show:      opts.Bool(true),
				Formatter: "{b} ({d}%)",
			}),
			charts.WithPieChartOpts(opts.PieChart{
				Radius:   []string{"30%", "75%"},
				RoseType: "radius",
			}),
		)

	return pie
}

// WorldMap generates a world map chart for country distribution.
func WorldMap(countries map[string]uint64) *charts.Map {
	// Calculate series data for the chart.
	items := make([]opts.MapData, 0, len(countries))
	maxVisits := uint64(0)
	for k, v := range countries {
		items = append(items, opts.MapData{Name: k, Value: v})
		if v > maxVisits {
			maxVisits = v
		}
	}

	mc := charts.NewMap()
	mc.RegisterMapType("world")
	mc.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Visits by Country",
		}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Calculable: opts.Bool(true),
			Min:        0,
			Max:        float32(maxVisits),
		}),
	)

	mc.AddSeries("Visits", items)

	return mc
}
