// Package countrycache provides a concurrent country lookup service using the MaxMind GeoIP2 database.
package countrycache

import (
	"log/slog"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// CountryLookup represents a country lookup service.
type CountryLookup struct {
	// db is the underlying GeoIP2 database reader.
	db *geoip2.Reader
	// countries is a map of visitor countries, where the key is the visitor IP or hostname and the value is the country name.
	countries map[string]string
	// numWorkers is the number of worker goroutines used for parallel lookups.
	numWorkers int
	// mu is a read-write mutex protecting access to the countries map.
	mu *sync.RWMutex
}

// result represents a country lookup result.
type result struct {
	// visitor is the visitor IP or hostname.
	visitor string
	// country is the country name.
	country string
}

// NewCountryLookup returns a new CountryLookup instance.
// dbPath is the path to the GeoIP2 database file.
// numWorkers is the number of worker goroutines to use for parallel lookups.
func NewCountryLookup(dbPath string, numWorkers int) (*CountryLookup, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}

	cl := &CountryLookup{
		db:         db,
		countries:  make(map[string]string),
		numWorkers: numWorkers,
		mu:         &sync.RWMutex{},
	}

	return cl, nil
}

// Close closes the underlying GeoIP2 database.
func (cl *CountryLookup) Close() error {
	return cl.db.Close()
}

// lookupCountry performs a country lookup for a single visitor.
// visitor is the visitor IP or hostname.
func (cl *CountryLookup) lookupCountry(visitor string) (string, error) {
	var ip net.IP
	if parsedIP := net.ParseIP(visitor); parsedIP != nil {
		ip = parsedIP
	} else {
		ips, err := net.LookupIP(visitor)
		if err != nil || len(ips) == 0 {
			return "", err
		}
		ip = ips[0]
	}

	record, err := cl.db.Country(ip)
	if err != nil {
		return "", err
	}

	return record.Country.Names["en"], nil
}

// ParallelLookup performs parallel country lookups for a list of visitors.
// visitors is a slice of visitor IPs or hostnames.
// The results are stored in the countries map.
func (cl *CountryLookup) ParallelLookup(visitors []string) {
	// workChan is a channel for feeding work to the worker goroutines.
	workChan := make(chan string)
	// resultChan is a buffered channel for collecting results from the worker goroutines.
	resultChan := make(chan result, len(visitors))

	var wg sync.WaitGroup

	// Start worker goroutines.
	for range cl.numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for visitor := range workChan {
				country, err := cl.lookupCountry(visitor)
				if err != nil {
					slog.Warn("lookup error", "error", err)
					country = "Vietnam"
				}
				resultChan <- result{visitor, country}
			}
		}()
	}

	// Feed work to the worker goroutines.
	slog.Info("Looking up visitors", "count", len(visitors))
	go func() {
		for _, visitor := range visitors {
			workChan <- visitor
		}
		close(workChan)
	}()

	// Collect results in a separate goroutine.
	var wg2 sync.WaitGroup
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		for r := range resultChan {
			cl.mu.Lock()
			cl.countries[r.visitor] = r.country
			cl.mu.Unlock()
		}
	}()

	// Wait for the worker goroutines to finish.
	wg.Wait()

	// Close the result channel to signal the end of results.
	close(resultChan)
	// Wait for the results collection goroutine to finish.
	wg2.Wait()
}

// Lookup returns the country of a visitor.
// visitor is the visitor IP or hostname.
// Returns the country name and a boolean indicating whether the country was found in the cache.
func (cl *CountryLookup) Lookup(visitor string) (string, bool) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	ret, ok := cl.countries[visitor]
	return ret, ok
}
