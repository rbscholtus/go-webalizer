package logstats

import "time"

// Stats holds aggregated metrics parsed from web server log files.
type HB struct {
	Hits  uint64
	Bytes uint64
}

func (h *HB) AddTraffic(bytes uint64) {
	h.Hits++
	h.Bytes += bytes
}

type HBV struct {
	Hits   uint64
	Bytes  uint64
	Visits uint64
}

func (h *HBV) AddTraffic(bytes uint64, isNewVisit bool) {
	h.Hits++
	h.Bytes += bytes
	if isNewVisit {
		h.Visits++
	}
}

type LogStats struct {
	Hits       map[string]uint64                    // Key: "YYYY-MM-DD"
	Files      map[string]uint64                    // Key: "YYYY-MM-DD"
	Pages      map[string]uint64                    // Key: "YYYY-MM-DD"
	Bytes      map[string]uint64                    // Key: "YYYY-MM-DD"
	Visits     map[string]map[string]uint64         // Key: "YYYY-MM-DD", value: map[IP]uint64
	FirstVisit map[string]time.Time                 // Key: IP
	LastVisit  map[string]time.Time                 // Key: IP
	Sites      map[string]map[string]uint64         // Key: "YYYY-MM-DD", value: map[IP]uint64
	Methods    map[string]map[string]uint64         // Key: "YYYY-MM-DD", value: map[Method]uint64
	RespCodes  map[string]map[uint16]uint64         // Key: "YYYY-MM-DD", value: map[ResponseCode]uint64
	IPs        map[string]map[string]*HBV           // Key: "YYYY-MM-DD", value: map[IP]*HBV
	UserAgents map[string]map[string]*HBV           // Key: "YYYY-MM-DD", value: map[UserAgent]*HBV
	URLPaths   map[string]map[string]map[string]*HB // Key: "YYYY-MM-DD", value: map[URLPath], value: map[Method]*HB
	Referrers  map[string]map[string]*HB            // Key: "YYYY-MM-DD", value: map[Referrer]*HB
}

func NewLogStats() *LogStats {
	return &LogStats{
		Hits:       make(map[string]uint64),
		Files:      make(map[string]uint64),
		Pages:      make(map[string]uint64),
		Bytes:      make(map[string]uint64),
		Visits:     make(map[string]map[string]uint64),
		FirstVisit: make(map[string]time.Time),
		LastVisit:  make(map[string]time.Time),
		Sites:      make(map[string]map[string]uint64),
		Methods:    make(map[string]map[string]uint64),
		RespCodes:  make(map[string]map[uint16]uint64),
		IPs:        make(map[string]map[string]*HBV),
		UserAgents: make(map[string]map[string]*HBV),
		URLPaths:   make(map[string]map[string]map[string]*HB),
		Referrers:  make(map[string]map[string]*HB),
	}
}

func (stats *LogStats) UpdateIPStats(date string, ip string, bytes uint64, isNewVisit bool) {
	if stats.IPs[date] == nil {
		stats.IPs[date] = make(map[string]*HBV)
	}
	if _, ok := stats.IPs[date][ip]; !ok {
		stats.IPs[date][ip] = &HBV{}
	}
	stats.IPs[date][ip].AddTraffic(bytes, isNewVisit)
}

func (stats *LogStats) UpdateUserAgentStats(date string, userAgent string, bytes uint64, isNewVisit bool) {
	if stats.UserAgents[date] == nil {
		stats.UserAgents[date] = make(map[string]*HBV)
	}
	if _, ok := stats.UserAgents[date][userAgent]; !ok {
		stats.UserAgents[date][userAgent] = &HBV{}
	}
	stats.UserAgents[date][userAgent].AddTraffic(bytes, isNewVisit)
}

func (stats *LogStats) UpdateURLStats(date string, URLPath string, method string, bytes uint64) {
	if stats.URLPaths[date] == nil {
		stats.URLPaths[date] = make(map[string]map[string]*HB)
	}
	if _, ok := stats.URLPaths[date][URLPath]; !ok {
		stats.URLPaths[date][URLPath] = make(map[string]*HB)
	}
	if _, ok := stats.URLPaths[date][URLPath][method]; !ok {
		stats.URLPaths[date][URLPath][method] = &HB{}
	}
	stats.URLPaths[date][URLPath][method].AddTraffic(bytes)
}

func (stats *LogStats) UpdateReferrerStats(date string, Referrer string, bytes uint64) {
	if stats.Referrers[date] == nil {
		stats.Referrers[date] = make(map[string]*HB)
	}
	if _, ok := stats.Referrers[date][Referrer]; !ok {
		stats.Referrers[date][Referrer] = &HB{}
	}
	stats.Referrers[date][Referrer].AddTraffic(bytes)
}

type MonthData struct {
	Month  string
	Hits   uint64
	Files  uint64
	Pages  uint64
	Bytes  uint64
	Visits uint64
	Sites  uint64
}

func (stats *LogStats) AggregatesByMonth() *map[string]*MonthData {
	aggr := make(map[string]*MonthData)
	for dateStr, hits := range stats.Hits {
		files := stats.Files[dateStr]
		pages := stats.Pages[dateStr]
		bytes := stats.Bytes[dateStr]
		visits := uint64(0)
		for _, count := range stats.Visits[dateStr] {
			visits += count
		}
		sites := uint64(len(stats.Sites[dateStr]))

		monthStr := dateStr[:7]
		value, ok := aggr[monthStr]
		if !ok {
			date, _ := time.Parse("2006-01", monthStr)
			value = &MonthData{date.Format("Jan"), hits, files, pages, bytes, visits, sites}
			aggr[monthStr] = value
		} else {
			value.Hits += hits
			value.Files += files
			value.Pages += pages
			value.Bytes += bytes
			value.Visits += visits
			value.Sites += sites
		}
	}

	return &aggr
}
