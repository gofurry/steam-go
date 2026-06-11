package steam

import (
	"time"

	itraffic "github.com/gofurry/steam-go/internal/traffic"
)

// RuntimeStats is one sanitized read-only snapshot of SDK runtime state.
type RuntimeStats struct {
	GeneratedAt time.Time
	Classes     map[TrafficClass]TrafficClassRuntimeStats
	Proxies     []ProxyMetricsSnapshot
}

// TrafficClassRuntimeStats groups runtime counters for one traffic class.
type TrafficClassRuntimeStats struct {
	Cache     CacheRuntimeStats
	Transport TransportRuntimeStats
}

// CacheRuntimeStats describes one in-memory cache runtime.
type CacheRuntimeStats struct {
	Entries         int
	MaxEntries      int
	Hits            uint64
	Misses          uint64
	Stores          uint64
	Evictions       uint64
	ConditionalHits uint64
}

// TransportRuntimeStats describes request-control state for one transport runtime.
type TransportRuntimeStats struct {
	HostControlKeys    int
	SessionControlKeys int
	HostWaits          uint64
	SessionWaits       uint64
	HostPrunes         uint64
	SessionPrunes      uint64
}

// RuntimeStats returns one sanitized snapshot of cache, request-control, and proxy runtime counters.
func (c *Client) RuntimeStats() RuntimeStats {
	stats := RuntimeStats{
		GeneratedAt: time.Now(),
		Classes:     make(map[TrafficClass]TrafficClassRuntimeStats),
	}
	if c == nil {
		return stats
	}

	for class, source := range c.rawRuntimes.statsSources {
		class = itraffic.NormalizeClass(class)
		classStats := TrafficClassRuntimeStats{}
		if source.cache != nil {
			cache := source.cache.Stats()
			classStats.Cache = CacheRuntimeStats{
				Entries:         cache.Entries,
				MaxEntries:      cache.MaxEntries,
				Hits:            cache.Hits,
				Misses:          cache.Misses,
				Stores:          cache.Stores,
				Evictions:       cache.Evictions,
				ConditionalHits: cache.ConditionalHits,
			}
		}
		if source.transport != nil {
			transport := source.transport.Stats()
			classStats.Transport = TransportRuntimeStats{
				HostControlKeys:    transport.HostControlKeys,
				SessionControlKeys: transport.SessionControlKeys,
				HostWaits:          transport.HostWaits,
				SessionWaits:       transport.SessionWaits,
				HostPrunes:         transport.HostPrunes,
				SessionPrunes:      transport.SessionPrunes,
			}
		}
		stats.Classes[class] = classStats
	}

	for _, provider := range c.rawRuntimes.proxyMetrics {
		if provider == nil {
			continue
		}
		stats.Proxies = append(stats.Proxies, provider.ProxyMetricsSnapshot())
	}
	return stats
}
