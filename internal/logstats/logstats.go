// Package logstats provides a data structure for holding and aggregating web server log statistics.
package logstats

import (
	"time"

	"github.com/rbscholtus/go-webalizer/internal/countrycache"
)

// HitsBytes holds aggregated metrics for hits and bytes.
type HitsBytes struct {
	// Hits is the total number of hits.
	Hits uint64
	// Bytes is the total number of bytes transferred.
	Bytes uint64
}

// AddTraffic increments the hits and bytes counters.
func (hb *HitsBytes) AddTraffic(bytes uint64) {
	hb.Hits++
	hb.Bytes += bytes
}

// HitsBytesVisits holds aggregated metrics for hits, bytes, and visits.
type HitsBytesVisits struct {
	// Hits is the total number of hits.
	Hits uint64
	// Bytes is the total number of bytes transferred.
	Bytes uint64
	// Visits is the total number of visits.
	Visits uint64
}

// AddTraffic increments the hits, bytes, and visits counters.
func (hbv *HitsBytesVisits) AddTraffic(bytes uint64, isNewVisit bool) {
	hbv.Hits++
	hbv.Bytes += bytes
	if isNewVisit {
		hbv.Visits++
	}
}

// LogStats holds aggregated metrics parsed from web server log files.
type LogStats struct {
	// Hits is a map of hits per day, keyed by date string in the format "YYYY-MM-DD".
	Hits map[string]uint64
	// Files is a map of file requests per day, keyed by date string in the format "YYYY-MM-DD".
	Files map[string]uint64
	// Pages is a map of page requests per day, keyed by date string in the format "YYYY-MM-DD".
	Pages map[string]uint64
	// Bytes is a map of bytes transferred per day, keyed by date string in the format "YYYY-MM-DD".
	Bytes map[string]uint64
	// Visits is a map of visits per day, keyed by date string in the format "YYYY-MM-DD" and IP address.
	Visits map[string]map[string]uint64
	// CtrVisits is a map of visits per day, keyed by date string in the format "YYYY-MM-DD" and country.
	CtrVisits map[string]map[string]uint64
	// FirstVisit is a map of first visit timestamps, keyed by IP address.
	FirstVisit map[string]time.Time
	// LastVisit is a map of last visit timestamps, keyed by IP address.
	LastVisit map[string]time.Time
	// Sites is a map of sites per day, keyed by date string in the format "YYYY-MM-DD" and IP address.
	Sites map[string]map[string]uint64
	// Methods is a map of HTTP methods per day, keyed by date string in the format "YYYY-MM-DD" and method.
	Methods map[string]map[string]uint64
	// RespCodes is a map of HTTP response codes per day, keyed by date string in the format "YYYY-MM-DD" and response code.
	RespCodes map[string]map[uint16]uint64
	// IPs is a map of IP statistics per day, keyed by date string in the format "YYYY-MM-DD" and IP address.
	IPs map[string]map[string]*HitsBytesVisits
	// UserAgents is a map of user agent statistics per day, keyed by date string in the format "YYYY-MM-DD" and user agent.
	UserAgents map[string]map[string]*HitsBytesVisits
	// URLPaths is a map of URL path statistics per day, keyed by date string in the format "YYYY-MM-DD", URL path, and method.
	URLPaths map[string]map[string]map[string]*HitsBytes
	// Referrers is a map of referrer statistics per day, keyed by date string in the format "YYYY-MM-DD" and referrer.
	Referrers map[string]map[string]*HitsBytes
}

// NewLogStats returns a new LogStats instance.
func NewLogStats() *LogStats {
	return &LogStats{
		Hits:       make(map[string]uint64),
		Files:      make(map[string]uint64),
		Pages:      make(map[string]uint64),
		Bytes:      make(map[string]uint64),
		Visits:     make(map[string]map[string]uint64),
		CtrVisits:  make(map[string]map[string]uint64),
		FirstVisit: make(map[string]time.Time),
		LastVisit:  make(map[string]time.Time),
		Sites:      make(map[string]map[string]uint64),
		Methods:    make(map[string]map[string]uint64),
		RespCodes:  make(map[string]map[uint16]uint64),
		IPs:        make(map[string]map[string]*HitsBytesVisits),
		UserAgents: make(map[string]map[string]*HitsBytesVisits),
		URLPaths:   make(map[string]map[string]map[string]*HitsBytes),
		Referrers:  make(map[string]map[string]*HitsBytes),
	}
}

// UpdateIPStats updates the IP statistics for a given date and IP address.
func (stats *LogStats) UpdateIPStats(date string, ip string, bytes uint64, isNewVisit bool) {
	if stats.IPs[date] == nil {
		stats.IPs[date] = make(map[string]*HitsBytesVisits)
	}
	if _, ok := stats.IPs[date][ip]; !ok {
		stats.IPs[date][ip] = &HitsBytesVisits{}
	}
	stats.IPs[date][ip].AddTraffic(bytes, isNewVisit)
}

// UpdateUserAgentStats updates the user agent statistics for a given date and user agent.
func (stats *LogStats) UpdateUserAgentStats(date string, userAgent string, bytes uint64, isNewVisit bool) {
	if stats.UserAgents[date] == nil {
		stats.UserAgents[date] = make(map[string]*HitsBytesVisits)
	}
	if _, ok := stats.UserAgents[date][userAgent]; !ok {
		stats.UserAgents[date][userAgent] = &HitsBytesVisits{}
	}
	stats.UserAgents[date][userAgent].AddTraffic(bytes, isNewVisit)
}

// UpdateURLStats updates the URL path statistics for a given date, URL path, and method.
func (stats *LogStats) UpdateURLStats(date string, URLPath string, method string, bytes uint64) {
	if stats.URLPaths[date] == nil {
		stats.URLPaths[date] = make(map[string]map[string]*HitsBytes)
	}
	if _, ok := stats.URLPaths[date][URLPath]; !ok {
		stats.URLPaths[date][URLPath] = make(map[string]*HitsBytes)
	}
	if _, ok := stats.URLPaths[date][URLPath][method]; !ok {
		stats.URLPaths[date][URLPath][method] = &HitsBytes{}
	}
	stats.URLPaths[date][URLPath][method].AddTraffic(bytes)
}

// UpdateReferrerStats updates the referrer statistics for a given date and referrer.
func (stats *LogStats) UpdateReferrerStats(date string, Referrer string, bytes uint64) {
	if stats.Referrers[date] == nil {
		stats.Referrers[date] = make(map[string]*HitsBytes)
	}
	if _, ok := stats.Referrers[date][Referrer]; !ok {
		stats.Referrers[date][Referrer] = &HitsBytes{}
	}
	stats.Referrers[date][Referrer].AddTraffic(bytes)
}

// uniqueVisitors returns a list of unique visitor IP addresses.
func uniqueVisitors(visitorsByDate map[string]map[string]uint64) []string {
	var keys []string
	seen := make(map[string]struct{})
	for _, visitors := range visitorsByDate {
		for visitor := range visitors {
			if _, exists := seen[visitor]; exists {
				continue
			}
			seen[visitor] = struct{}{}
			keys = append(keys, visitor)
		}
	}
	return keys
}

// LookupCountries performs a country lookup for all unique visitors and updates the CtrVisits map.
func (stats *LogStats) LookupCountries() error {
	// Create a new country cache instance.
	cl, err := countrycache.NewCountryLookup("./GeoLite2-Country.mmdb", 32)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Get a list of unique visitor IP addresses.
	visitors := uniqueVisitors(stats.Visits)
	// Perform a parallel country lookup for all unique visitors.
	cl.ParallelLookup(visitors)

	// Update the CtrVisits map with the country lookup results.
	for date, ipMaps := range stats.Visits {
		for visitor, visits := range ipMaps {
			if stats.CtrVisits[date] == nil {
				stats.CtrVisits[date] = make(map[string]uint64)
			}

			// Get the country name for the visitor IP address.
			if country, ok := cl.Lookup(visitor); ok {
				// Increment the visit count for the country.
				stats.CtrVisits[date][country] += visits
			}
		}
	}

	return nil
}

// HFPBVSData holds aggregated metrics for hits, files, pages, bytes, visits, and sites.
type HFPBVSData struct {
	// Category is the category name (e.g. month name).
	Category string
	// Hits is the total number of hits.
	Hits uint64
	// Files is the total number of file requests.
	Files uint64
	// Pages is the total number of page requests.
	Pages uint64
	// Bytes is the total number of bytes transferred.
	Bytes uint64
	// Visits is the total number of visits.
	Visits uint64
	// Sites is the total number of sites.
	Sites uint64
}

// CategoryData holds a category and its corresponding count.
type CategoryData struct {
	// Category is the name of the category.
	Category string
	// Count is the number of items in the category.
	Count uint64
}

// AggregatesByMonth returns a map of aggregated metrics by month.
func (stats *LogStats) AggregatesByMonth() map[string]*HFPBVSData {
	aggr := make(map[string]*HFPBVSData)
	for dateStr, hits := range stats.Hits {
		// Get the file, page, byte, and visit counts for the date.
		files := stats.Files[dateStr]
		pages := stats.Pages[dateStr]
		bytes := stats.Bytes[dateStr]
		visits := uint64(0)
		for _, count := range stats.Visits[dateStr] {
			visits += count
		}
		sites := uint64(len(stats.Sites[dateStr]))

		// Get the month string from the date string.
		monthStr := dateStr[:7]
		value, ok := aggr[monthStr]
		if !ok {
			// Create a new HFPBVSData instance for the month if it doesn't exist.
			date, _ := time.Parse("2006-01", monthStr)
			value = &HFPBVSData{date.Format("Jan"), hits, files, pages, bytes, visits, sites}
			aggr[monthStr] = value
		} else {
			// Increment the metrics for the month.
			value.Hits += hits
			value.Files += files
			value.Pages += pages
			value.Bytes += bytes
			value.Visits += visits
			value.Sites += sites
		}
	}

	return aggr
}

// recentKeys returns a list of date strings for the last month.
func (stats *LogStats) recentKeys() []string {
	// Find the last date in the stats.
	var lastKey string
	for key := range stats.Hits {
		if key > lastKey {
			lastKey = key
		}
	}

	// Determine the date range for the last month.
	lastTime, _ := time.Parse("2006-01-02", lastKey)
	lastTime = lastTime.AddDate(0, 0, 1)
	firstTime := lastTime.AddDate(0, -1, 0)
	firstKey, lastKey := firstTime.Format("2006-01-02"), lastTime.Format("2006-01-02")

	// Get the date strings for the last month.
	daysKeys := make([]string, 0, 31)
	for key := range stats.Hits {
		if key >= firstKey && key < lastKey {
			daysKeys = append(daysKeys, key)
		}
	}

	return daysKeys
}

// RecentAggregates returns a map of aggregated metrics for the last month.
func (stats *LogStats) RecentAggregates() map[string]*HFPBVSData {
	daysKeys := stats.recentKeys()

	aggr := make(map[string]*HFPBVSData, 31)
	for _, dateStr := range daysKeys {
		t, _ := time.Parse("2006-01-02", dateStr)
		formattedDate := t.Format("Jan 2")

		aggr[dateStr] = &HFPBVSData{
			formattedDate,
			stats.Hits[dateStr],
			stats.Files[dateStr],
			stats.Pages[dateStr],
			stats.Bytes[dateStr],
			uint64(0),
			uint64(len(stats.Sites[dateStr])),
		}
		for _, count := range stats.Visits[dateStr] {
			aggr[dateStr].Visits += count
		}
	}

	return aggr
}

// MethRespAggregates returns maps of aggregated metrics for HTTP methods and response codes.
func (stats *LogStats) MethRespAggregates() (map[string]uint64, map[uint16]uint64) {
	daysKeys := stats.recentKeys()

	aggr := make(map[string]uint64)
	aggr2 := make(map[uint16]uint64)
	for _, date := range daysKeys {
		for meth, hits := range stats.Methods[date] {
			if meth != "-" {
				aggr[meth] += hits
			}
		}
		for resp, hits := range stats.RespCodes[date] {
			aggr2[resp] += hits
		}
	}

	return aggr, aggr2
}

// CountryAggregates returns a map of aggregated metrics for countries.
func (stats *LogStats) CountryAggregates() map[string]uint64 {
	daysKeys := stats.recentKeys()

	aggr := make(map[string]uint64)
	for _, date := range daysKeys {
		for visitor, hits := range stats.CtrVisits[date] {
			aggr[visitor] += hits
		}
	}

	return aggr
}
