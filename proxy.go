package steam

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

	itraffic "github.com/GoFurry/steam-go/internal/traffic"
	itransport "github.com/GoFurry/steam-go/internal/transport"
)

const defaultProxyClientTimeout = 10 * time.Second

var ErrAllProxiesCoolingDown = errors.New("all proxies are cooling down")

const (
	stickyProxyMaxEntries       = 512
	stickyProxyIdleTTL          = 10 * time.Minute
	stickyProxySweepIntervalOps = 128
)

type proxyHealthStatusReporter interface {
	ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
}

type proxyAvailabilityChecker interface {
	proxyAvailable(proxyURL *url.URL, now time.Time) bool
}

type proxySessionKeyContextKey struct{}

// ProxyMetricsProvider exposes one read-only snapshot of proxy pool health metrics.
type ProxyMetricsProvider interface {
	ProxyMetricsSnapshot() ProxyMetricsSnapshot
}

// ProxyHealthConfig configures one health-checked proxy pool.
type ProxyHealthConfig struct {
	FailureThreshold int
	Cooldown         time.Duration
}

// ProxyMetricsSnapshot is one immutable view of a proxy pool at one point in time.
type ProxyMetricsSnapshot struct {
	GeneratedAt    time.Time
	TotalProxies   int
	HealthyProxies int
	CoolingProxies int
	Proxies        []ProxyEndpointMetrics
}

// ProxyEndpointMetrics describes one proxy endpoint inside a health-checked pool.
type ProxyEndpointMetrics struct {
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

// DefaultProxyHealthConfig returns the default proxy health settings.
func DefaultProxyHealthConfig() ProxyHealthConfig {
	return ProxyHealthConfig{
		FailureThreshold: 2,
		Cooldown:         30 * time.Second,
	}
}

// WithProxySessionKey attaches one explicit sticky-proxy session key to a request context.
//
// Blank keys are treated as "unset" and leave the context unchanged.
// A nil context falls back to context.Background().
func WithProxySessionKey(ctx context.Context, key string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return ctx
	}
	return context.WithValue(ctx, proxySessionKeyContextKey{}, key)
}

// NewStaticProxySelector creates a selector that always returns the same proxy.
//
// Empty strings are treated as "no proxy" and return nil, nil.
func NewStaticProxySelector(rawURL string) (ProxySelector, error) {
	urls, err := parseProxyURLs(rawURL)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, nil
	}
	return staticProxySelector{url: urls[0]}, nil
}

// NewRoundRobinProxySelector creates a selector that rotates across proxies.
//
// Empty strings are ignored. If no valid proxies remain, nil, nil is returned.
func NewRoundRobinProxySelector(rawURLs ...string) (ProxySelector, error) {
	urls, err := parseProxyURLs(rawURLs...)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, nil
	}
	if len(urls) == 1 {
		return staticProxySelector{url: urls[0]}, nil
	}
	return &roundRobinProxySelector{urls: urls}, nil
}

// NewHealthCheckedRoundRobinProxySelector creates a round-robin proxy pool with failure scoring and cooldown.
func NewHealthCheckedRoundRobinProxySelector(cfg ProxyHealthConfig, rawURLs ...string) (ProxySelector, error) {
	urls, err := parseProxyURLs(rawURLs...)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, nil
	}

	resolved, err := resolveProxyHealthConfig(cfg)
	if err != nil {
		return nil, err
	}

	states := make([]proxyHealthState, len(urls))
	indexByProxy := make(map[string]int, len(urls))
	for i, proxyURL := range urls {
		indexByProxy[proxyURL.String()] = i
	}

	return &healthCheckedRoundRobinProxySelector{
		urls:          urls,
		states:        states,
		indexByProxy:  indexByProxy,
		failureThresh: resolved.FailureThreshold,
		cooldown:      resolved.Cooldown,
	}, nil
}

// NewStickyProxySelector wraps a base selector and keeps one proxy choice per explicit session key.
//
// When no session key is attached to the request context, it falls back to the base selector behavior.
func NewStickyProxySelector(base ProxySelector) ProxySelector {
	if base == nil {
		return nil
	}
	sticky := &stickyProxySelector{
		base:  base,
		cache: make(map[string]stickyProxyEntry),
	}
	if metrics, ok := base.(ProxyMetricsProvider); ok {
		return &stickyProxySelectorWithMetrics{
			stickyProxySelector: sticky,
			metrics:             metrics,
		}
	}
	return sticky
}

// ProxyRoute describes one host/path based proxy routing rule.
//
// The first matching route wins. An empty ProxyURL means "send direct".
type ProxyRoute struct {
	Host       string
	PathPrefix string
	ProxyURL   string
}

// NewRoutingProxySelector creates a selector that routes requests by host/path.
//
// Host matching is case-insensitive and compares against req.URL.Hostname().
// PathPrefix matches req.URL.Path via strings.HasPrefix.
// The first matching route wins; no match falls back to direct.
func NewRoutingProxySelector(routes ...ProxyRoute) (ProxySelector, error) {
	compiled := make([]compiledProxyRoute, 0, len(routes))
	for _, route := range routes {
		var parsed *url.URL
		if strings.TrimSpace(route.ProxyURL) != "" {
			urls, err := parseProxyURLs(route.ProxyURL)
			if err != nil {
				return nil, err
			}
			parsed = urls[0]
		}

		compiled = append(compiled, compiledProxyRoute{
			host:       strings.ToLower(strings.TrimSpace(route.Host)),
			pathPrefix: strings.TrimSpace(route.PathPrefix),
			proxyURL:   parsed,
		})
	}
	if len(compiled) == 0 {
		return nil, nil
	}
	return routingProxySelector{routes: compiled}, nil
}

// NewHTTPClientWithProxySelector builds an http.Client backed by one ProxySelector.
//
// Passing a nil selector keeps the client in direct mode.
// A zero timeout falls back to a safe default timeout.
func NewHTTPClientWithProxySelector(selector ProxySelector, timeout time.Duration) (*http.Client, error) {
	if timeout < 0 {
		return nil, fmt.Errorf("timeout must not be negative")
	}
	if timeout == 0 {
		timeout = defaultProxyClientTimeout
	}

	client, err := WrapHTTPClientWithProxySelector(&http.Client{
		Timeout:   timeout,
		Transport: http.DefaultTransport,
	}, selector)
	if err != nil {
		return nil, err
	}
	client.Timeout = timeout
	return client, nil
}

// WrapHTTPClientWithProxySelector clones one base client and applies proxy selection to the clone.
//
// This helper is useful inside ProxyAwareTransportHook implementations that need to customize the base client
// and then re-attach the SDK proxy selection wrapper.
func WrapHTTPClientWithProxySelector(base *http.Client, selector ProxySelector) (*http.Client, error) {
	if base == nil {
		return nil, fmt.Errorf("http client is required")
	}

	cloned := cloneHTTPClient(base)
	rt, err := itransport.WrapRoundTripper(cloned.Transport, selector)
	if err != nil {
		return nil, err
	}
	cloned.Transport = rt
	return cloned, nil
}

type staticProxySelector struct {
	url *url.URL
}

func (s staticProxySelector) Next(*http.Request) (*url.URL, error) {
	return s.url, nil
}

type roundRobinProxySelector struct {
	urls []*url.URL
	next atomic.Uint64
}

func (s *roundRobinProxySelector) Next(*http.Request) (*url.URL, error) {
	current := s.next.Add(1) - 1
	return cloneProxyURL(s.urls[current%uint64(len(s.urls))]), nil
}

type stickyProxySelector struct {
	base  ProxySelector
	mu    sync.RWMutex
	cache map[string]stickyProxyEntry

	opCount   atomic.Uint64
	lastSweep atomic.Int64
}

type stickyProxySelectorWithMetrics struct {
	*stickyProxySelector
	metrics ProxyMetricsProvider
}

type stickyProxyEntry struct {
	proxyURL *url.URL
	lastUsed time.Time
}

func (s *stickyProxySelector) Next(req *http.Request) (*url.URL, error) {
	key := proxySessionKeyFromRequest(req)
	if key == "" {
		return s.base.Next(req)
	}

	now := time.Now()
	s.maybePrune(now)

	s.mu.RLock()
	cached, ok := s.cache[key]
	s.mu.RUnlock()
	if ok {
		if checker, ok := s.base.(proxyAvailabilityChecker); ok && !checker.proxyAvailable(cached.proxyURL, now) {
			s.mu.Lock()
			delete(s.cache, key)
			s.mu.Unlock()
		} else {
			s.markUsed(key, now)
			return cloneProxyURL(cached.proxyURL), nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cached, ok = s.cache[key]
	if ok {
		if checker, ok := s.base.(proxyAvailabilityChecker); ok && !checker.proxyAvailable(cached.proxyURL, now) {
			delete(s.cache, key)
		} else {
			cached.lastUsed = now
			s.cache[key] = cached
			return cloneProxyURL(cached.proxyURL), nil
		}
	}

	proxyURL, err := s.base.Next(req)
	if err != nil {
		return nil, err
	}
	s.cache[key] = stickyProxyEntry{
		proxyURL: cloneProxyURL(proxyURL),
		lastUsed: now,
	}
	if len(s.cache) > stickyProxyMaxEntries {
		s.pruneLocked(now, len(s.cache)-stickyProxyMaxEntries)
	}
	return cloneProxyURL(proxyURL), nil
}

func (s *stickyProxySelector) ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error) {
	if reporter, ok := s.base.(proxyHealthStatusReporter); ok {
		reporter.ReportProxyResult(req, proxyURL, statusCode, err)
	}
}

func (s *stickyProxySelectorWithMetrics) ProxyMetricsSnapshot() ProxyMetricsSnapshot {
	return s.metrics.ProxyMetricsSnapshot()
}

func (s *stickyProxySelector) markUsed(key string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.cache[key]
	if !ok {
		return
	}
	entry.lastUsed = now
	s.cache[key] = entry
}

func (s *stickyProxySelector) maybePrune(now time.Time) {
	if s == nil {
		return
	}
	count := s.opCount.Add(1)
	if count%stickyProxySweepIntervalOps != 0 {
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

func (s *stickyProxySelector) pruneLocked(now time.Time, targetExtra int) {
	if len(s.cache) == 0 {
		return
	}

	idleBefore := now.Add(-stickyProxyIdleTTL)
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

	if targetExtra <= 0 || len(s.cache) <= stickyProxyMaxEntries {
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

type proxyHealthState struct {
	failureScore   int
	cooldownUntil  time.Time
	selectionCount uint64
	successCount   uint64
	failureCount   uint64
	cooldownCount  uint64
	lastFailureAt  time.Time
	lastSuccessAt  time.Time
}

type healthCheckedRoundRobinProxySelector struct {
	mu            sync.Mutex
	urls          []*url.URL
	states        []proxyHealthState
	indexByProxy  map[string]int
	failureThresh int
	cooldown      time.Duration
	nextIndex     int
}

func (s *healthCheckedRoundRobinProxySelector) Next(*http.Request) (*url.URL, error) {
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
		return cloneProxyURL(s.urls[idx]), nil
	}
	return nil, ErrAllProxiesCoolingDown
}

func (s *healthCheckedRoundRobinProxySelector) ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error) {
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
	if proxyResultFailedForRequest(req, statusCode, err) {
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

func (s *healthCheckedRoundRobinProxySelector) proxyAvailable(proxyURL *url.URL, now time.Time) bool {
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

func (s *healthCheckedRoundRobinProxySelector) ProxyMetricsSnapshot() ProxyMetricsSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	snapshot := ProxyMetricsSnapshot{
		GeneratedAt:  now,
		TotalProxies: len(s.urls),
		Proxies:      make([]ProxyEndpointMetrics, 0, len(s.urls)),
	}
	for i, proxyURL := range s.urls {
		state := s.states[i]
		if state.cooldownUntil.IsZero() || !now.Before(state.cooldownUntil) {
			snapshot.HealthyProxies++
		} else {
			snapshot.CoolingProxies++
		}
		snapshot.Proxies = append(snapshot.Proxies, ProxyEndpointMetrics{
			ProxyURL:       redactedProxyURLString(proxyURL),
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

type compiledProxyRoute struct {
	host       string
	pathPrefix string
	proxyURL   *url.URL
}

type routingProxySelector struct {
	routes []compiledProxyRoute
}

func (s routingProxySelector) Next(req *http.Request) (*url.URL, error) {
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

func parseProxyURLs(rawURLs ...string) ([]*url.URL, error) {
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

func resolveProxyHealthConfig(cfg ProxyHealthConfig) (ProxyHealthConfig, error) {
	resolved := DefaultProxyHealthConfig()

	switch {
	case cfg.FailureThreshold < 0:
		return ProxyHealthConfig{}, fmt.Errorf("proxy failure threshold must not be negative")
	case cfg.FailureThreshold > 0:
		resolved.FailureThreshold = cfg.FailureThreshold
	}

	switch {
	case cfg.Cooldown < 0:
		return ProxyHealthConfig{}, fmt.Errorf("proxy cooldown must not be negative")
	case cfg.Cooldown > 0:
		resolved.Cooldown = cfg.Cooldown
	}

	return resolved, nil
}

func proxySessionKeyFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	value, _ := req.Context().Value(proxySessionKeyContextKey{}).(string)
	return strings.TrimSpace(value)
}

func cloneProxyURL(proxyURL *url.URL) *url.URL {
	if proxyURL == nil {
		return nil
	}
	cloned := *proxyURL
	return &cloned
}

func redactedProxyURLString(proxyURL *url.URL) string {
	if proxyURL == nil {
		return ""
	}
	cloned := cloneProxyURL(proxyURL)
	cloned.User = nil
	return cloned.String()
}

func proxyResultFailed(statusCode int, err error) bool {
	if err != nil {
		return true
	}
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}

func proxyResultFailedForRequest(req *http.Request, statusCode int, err error) bool {
	if proxyResultFailed(statusCode, err) {
		return true
	}
	if statusCode != http.StatusForbidden || req == nil {
		return false
	}
	class, ok := itraffic.ClassFromContext(req.Context())
	return ok && class == itraffic.ClassPublicStorePage && itraffic.BlockDetectionFromContext(req.Context())
}
