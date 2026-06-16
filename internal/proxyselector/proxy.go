package proxyselector

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	itraffic "github.com/gofurry/steam-go/internal/traffic"
)

var ErrAllCoolingDown = errors.New("all proxies are cooling down")

const (
	stickyMaxEntries       = 512
	stickyIdleTTL          = 10 * time.Minute
	stickySweepIntervalOps = 128
)

type sessionKeyContextKey struct{}

// Selector chooses a proxy URL for a request.
type Selector interface {
	Next(req *http.Request) (*url.URL, error)
}

// ResultReporter records one proxy request outcome.
type ResultReporter interface {
	ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
}

// AvailabilityChecker reports whether a cached proxy can still be reused.
type AvailabilityChecker interface {
	ProxyAvailable(proxyURL *url.URL, now time.Time) bool
}

// MetricsProvider exposes one read-only internal snapshot of proxy pool health metrics.
type MetricsProvider interface {
	MetricsSnapshot() MetricsSnapshot
}

// HealthConfig configures one health-checked proxy pool.
type HealthConfig struct {
	FailureThreshold int
	Cooldown         time.Duration
}

// MetricsSnapshot is one immutable view of a proxy pool at one point in time.
type MetricsSnapshot struct {
	GeneratedAt    time.Time
	TotalProxies   int
	HealthyProxies int
	CoolingProxies int
	Proxies        []EndpointMetrics
}

// EndpointMetrics describes one proxy endpoint inside a health-checked pool.
type EndpointMetrics struct {
	ProxyURL       string
	FailureScore   int
	CooldownUntil  time.Time
	SelectionCount uint64
	SuccessCount   uint64
	FailureCount   uint64
	CooldownCount  uint64
	LastFailureAt  time.Time
	LastSuccessAt  time.Time
}

// DefaultHealthConfig returns the default proxy health settings.
func DefaultHealthConfig() HealthConfig {
	return HealthConfig{
		FailureThreshold: 2,
		Cooldown:         30 * time.Second,
	}
}

// WithSessionKey attaches one explicit sticky-proxy session key to a request context.
func WithSessionKey(ctx context.Context, key string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return ctx
	}
	return context.WithValue(ctx, sessionKeyContextKey{}, key)
}

// SessionKeyFromRequest returns the sticky proxy session key attached to a request.
func SessionKeyFromRequest(req *http.Request) string {
	return sessionKeyFromRequest(req)
}

// NewStaticSelector creates a selector that always returns the same proxy.
func NewStaticSelector(rawURL string) (Selector, error) {
	urls, err := parseURLs(rawURL)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, nil
	}
	return staticSelector{url: urls[0]}, nil
}

// NewRoundRobinSelector creates a selector that rotates across proxies.
func NewRoundRobinSelector(rawURLs ...string) (Selector, error) {
	urls, err := parseURLs(rawURLs...)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, nil
	}
	if len(urls) == 1 {
		return staticSelector{url: urls[0]}, nil
	}
	return &roundRobinSelector{urls: urls}, nil
}

// NewHealthCheckedRoundRobinSelector creates a round-robin proxy pool with failure scoring and cooldown.
func NewHealthCheckedRoundRobinSelector(cfg HealthConfig, rawURLs ...string) (Selector, error) {
	urls, err := parseURLs(rawURLs...)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, nil
	}

	resolved, err := resolveHealthConfig(cfg)
	if err != nil {
		return nil, err
	}

	states := make([]healthState, len(urls))
	indexByProxy := make(map[string]int, len(urls))
	for i, proxyURL := range urls {
		indexByProxy[proxyURL.String()] = i
	}

	return &healthCheckedRoundRobinSelector{
		urls:          urls,
		states:        states,
		indexByProxy:  indexByProxy,
		failureThresh: resolved.FailureThreshold,
		cooldown:      resolved.Cooldown,
	}, nil
}

// NewStickySelector wraps a base selector and keeps one proxy choice per explicit session key.
func NewStickySelector(base Selector) Selector {
	if base == nil {
		return nil
	}
	sticky := &stickySelector{
		base:  base,
		cache: make(map[string]stickyEntry),
	}
	if metrics, ok := base.(MetricsProvider); ok {
		return &stickySelectorWithMetrics{
			stickySelector: sticky,
			metrics:        metrics,
		}
	}
	return sticky
}

// Route describes one host/path based proxy routing rule.
type Route struct {
	Host       string
	PathPrefix string
	ProxyURL   string
}

// NewRoutingSelector creates a selector that routes requests by host/path.
func NewRoutingSelector(routes ...Route) (Selector, error) {
	compiled := make([]compiledRoute, 0, len(routes))
	for _, route := range routes {
		var parsed *url.URL
		if strings.TrimSpace(route.ProxyURL) != "" {
			urls, err := parseURLs(route.ProxyURL)
			if err != nil {
				return nil, err
			}
			parsed = urls[0]
		}

		compiled = append(compiled, compiledRoute{
			host:       strings.ToLower(strings.TrimSpace(route.Host)),
			pathPrefix: strings.TrimSpace(route.PathPrefix),
			proxyURL:   parsed,
		})
	}
	if len(compiled) == 0 {
		return nil, nil
	}
	return routingSelector{routes: compiled}, nil
}

type staticSelector struct {
	url *url.URL
}

func (s staticSelector) Next(*http.Request) (*url.URL, error) {
	return s.url, nil
}

type roundRobinSelector struct {
	urls []*url.URL
	next atomic.Uint64
}

func (s *roundRobinSelector) Next(*http.Request) (*url.URL, error) {
	current := s.next.Add(1) - 1
	return cloneURL(s.urls[current%uint64(len(s.urls))]), nil
}

type stickySelector struct {
	base  Selector
	mu    sync.RWMutex
	cache map[string]stickyEntry

	opCount   atomic.Uint64
	lastSweep atomic.Int64
}

type stickySelectorWithMetrics struct {
	*stickySelector
	metrics MetricsProvider
}

type stickyEntry struct {
	proxyURL *url.URL
	lastUsed time.Time
}

func (s *stickySelector) Next(req *http.Request) (*url.URL, error) {
	key := sessionKeyFromRequest(req)
	if key == "" {
		return s.base.Next(req)
	}

	now := time.Now()
	s.maybePrune(now)

	s.mu.RLock()
	cached, ok := s.cache[key]
	s.mu.RUnlock()
	if ok {
		if checker, ok := s.base.(AvailabilityChecker); ok && !checker.ProxyAvailable(cached.proxyURL, now) {
			s.mu.Lock()
			delete(s.cache, key)
			s.mu.Unlock()
		} else {
			s.markUsed(key, now)
			return cloneURL(cached.proxyURL), nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cached, ok = s.cache[key]
	if ok {
		if checker, ok := s.base.(AvailabilityChecker); ok && !checker.ProxyAvailable(cached.proxyURL, now) {
			delete(s.cache, key)
		} else {
			cached.lastUsed = now
			s.cache[key] = cached
			return cloneURL(cached.proxyURL), nil
		}
	}

	proxyURL, err := s.base.Next(req)
	if err != nil {
		return nil, err
	}
	s.cache[key] = stickyEntry{
		proxyURL: cloneURL(proxyURL),
		lastUsed: now,
	}
	if len(s.cache) > stickyMaxEntries {
		s.pruneLocked(now, len(s.cache)-stickyMaxEntries)
	}
	return cloneURL(proxyURL), nil
}

func (s *stickySelector) ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error) {
	if reporter, ok := s.base.(ResultReporter); ok {
		reporter.ReportProxyResult(req, proxyURL, statusCode, err)
	}
}

func (s *stickySelectorWithMetrics) MetricsSnapshot() MetricsSnapshot {
	return s.metrics.MetricsSnapshot()
}

func (s *stickySelector) markUsed(key string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.cache[key]
	if !ok {
		return
	}
	entry.lastUsed = now
	s.cache[key] = entry
}

func (s *stickySelector) maybePrune(now time.Time) {
	if s == nil {
		return
	}
	count := s.opCount.Add(1)
	if count%stickySweepIntervalOps != 0 {
		return
	}
	last := s.lastSweep.Load()
	if last != 0 && now.UnixNano()-last < int64(time.Second) {
		return
	}

	s.mu.Lock()
	s.pruneLocked(now, 0)
	s.mu.Unlock()
	s.lastSweep.Store(now.UnixNano())
}

func (s *stickySelector) pruneLocked(now time.Time, targetExtra int) {
	if len(s.cache) == 0 {
		return
	}

	idleBefore := now.Add(-stickyIdleTTL)
	type candidate struct {
		key      string
		lastUsed time.Time
	}

	candidates := make([]candidate, 0, len(s.cache))
	for key, entry := range s.cache {
		if entry.lastUsed.Before(idleBefore) {
			delete(s.cache, key)
			continue
		}
		candidates = append(candidates, candidate{
			key:      key,
			lastUsed: entry.lastUsed,
		})
	}

	if targetExtra <= 0 || len(s.cache) <= stickyMaxEntries {
		return
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].lastUsed.Before(candidates[j].lastUsed)
	})
	for _, candidate := range candidates {
		if targetExtra <= 0 {
			break
		}
		delete(s.cache, candidate.key)
		targetExtra--
	}
}

type healthState struct {
	failureScore   int
	cooldownUntil  time.Time
	selectionCount uint64
	successCount   uint64
	failureCount   uint64
	cooldownCount  uint64
	lastFailureAt  time.Time
	lastSuccessAt  time.Time
}

type healthCheckedRoundRobinSelector struct {
	mu            sync.Mutex
	urls          []*url.URL
	states        []healthState
	indexByProxy  map[string]int
	failureThresh int
	cooldown      time.Duration
	nextIndex     int
}

func (s *healthCheckedRoundRobinSelector) Next(*http.Request) (*url.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for i := 0; i < len(s.urls); i++ {
		idx := (s.nextIndex + i) % len(s.urls)
		if !s.states[idx].cooldownUntil.IsZero() && now.Before(s.states[idx].cooldownUntil) {
			continue
		}
		s.states[idx].selectionCount++
		s.nextIndex = (idx + 1) % len(s.urls)
		return cloneURL(s.urls[idx]), nil
	}
	return nil, ErrAllCoolingDown
}

func (s *healthCheckedRoundRobinSelector) ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error) {
	if proxyURL == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	idx, ok := s.indexByProxy[proxyURL.String()]
	if !ok {
		return
	}

	state := &s.states[idx]
	if resultFailedForRequest(req, statusCode, err) {
		state.failureCount++
		state.lastFailureAt = time.Now()
		state.failureScore++
		if state.failureScore >= s.failureThresh {
			state.failureScore = 0
			state.cooldownUntil = time.Now().Add(s.cooldown)
			state.cooldownCount++
		}
		return
	}

	state.successCount++
	state.lastSuccessAt = time.Now()
	state.failureScore = 0
	state.cooldownUntil = time.Time{}
}

func (s *healthCheckedRoundRobinSelector) ProxyAvailable(proxyURL *url.URL, now time.Time) bool {
	if proxyURL == nil {
		return true
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	idx, ok := s.indexByProxy[proxyURL.String()]
	if !ok {
		return true
	}
	return s.states[idx].cooldownUntil.IsZero() || !now.Before(s.states[idx].cooldownUntil)
}

func (s *healthCheckedRoundRobinSelector) MetricsSnapshot() MetricsSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	snapshot := MetricsSnapshot{
		GeneratedAt:  now,
		TotalProxies: len(s.urls),
		Proxies:      make([]EndpointMetrics, 0, len(s.urls)),
	}
	for i, proxyURL := range s.urls {
		state := s.states[i]
		if state.cooldownUntil.IsZero() || !now.Before(state.cooldownUntil) {
			snapshot.HealthyProxies++
		} else {
			snapshot.CoolingProxies++
		}
		snapshot.Proxies = append(snapshot.Proxies, EndpointMetrics{
			ProxyURL:       redactedURLString(proxyURL),
			FailureScore:   state.failureScore,
			CooldownUntil:  state.cooldownUntil,
			SelectionCount: state.selectionCount,
			SuccessCount:   state.successCount,
			FailureCount:   state.failureCount,
			CooldownCount:  state.cooldownCount,
			LastFailureAt:  state.lastFailureAt,
			LastSuccessAt:  state.lastSuccessAt,
		})
	}
	return snapshot
}

type compiledRoute struct {
	host       string
	pathPrefix string
	proxyURL   *url.URL
}

type routingSelector struct {
	routes []compiledRoute
}

func (s routingSelector) Next(req *http.Request) (*url.URL, error) {
	if req == nil || req.URL == nil {
		return nil, nil
	}

	host := strings.ToLower(req.URL.Hostname())
	path := req.URL.Path
	for _, route := range s.routes {
		if route.host != "" && route.host != host {
			continue
		}
		if route.pathPrefix != "" && !strings.HasPrefix(path, route.pathPrefix) {
			continue
		}
		return route.proxyURL, nil
	}
	return nil, nil
}

func parseURLs(rawURLs ...string) ([]*url.URL, error) {
	urls := make([]*url.URL, 0, len(rawURLs))
	for _, rawURL := range rawURLs {
		rawURL = strings.TrimSpace(rawURL)
		if rawURL == "" {
			continue
		}

		parsed, err := url.Parse(rawURL)
		if err != nil {
			return nil, fmt.Errorf("parse proxy url %q: %w", rawURL, err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return nil, fmt.Errorf("proxy url must include scheme and host")
		}
		urls = append(urls, parsed)
	}
	return urls, nil
}

func resolveHealthConfig(cfg HealthConfig) (HealthConfig, error) {
	resolved := DefaultHealthConfig()

	switch {
	case cfg.FailureThreshold < 0:
		return HealthConfig{}, fmt.Errorf("proxy failure threshold must not be negative")
	case cfg.FailureThreshold > 0:
		resolved.FailureThreshold = cfg.FailureThreshold
	}

	switch {
	case cfg.Cooldown < 0:
		return HealthConfig{}, fmt.Errorf("proxy cooldown must not be negative")
	case cfg.Cooldown > 0:
		resolved.Cooldown = cfg.Cooldown
	}

	return resolved, nil
}

func sessionKeyFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	value, _ := req.Context().Value(sessionKeyContextKey{}).(string)
	return strings.TrimSpace(value)
}

func cloneURL(proxyURL *url.URL) *url.URL {
	if proxyURL == nil {
		return nil
	}
	cloned := *proxyURL
	return &cloned
}

func redactedURLString(proxyURL *url.URL) string {
	if proxyURL == nil {
		return ""
	}
	cloned := cloneURL(proxyURL)
	cloned.User = nil
	return cloned.String()
}

func resultFailed(statusCode int, err error) bool {
	if err != nil {
		return true
	}
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}

func resultFailedForRequest(req *http.Request, statusCode int, err error) bool {
	if resultFailed(statusCode, err) {
		return true
	}
	if statusCode != http.StatusForbidden || req == nil {
		return false
	}
	class, ok := itraffic.ClassFromContext(req.Context())
	return ok && class == itraffic.ClassPublicStorePage && itraffic.BlockDetectionFromContext(req.Context())
}
