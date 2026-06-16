package steam

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gofurry/steam-go/internal/proxyselector"
	itransport "github.com/gofurry/steam-go/internal/transport"
)

const defaultProxyClientTimeout = 10 * time.Second

var ErrAllProxiesCoolingDown = errors.New("all proxies are cooling down")

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
	cfg := proxyselector.DefaultHealthConfig()
	return ProxyHealthConfig{
		FailureThreshold: cfg.FailureThreshold,
		Cooldown:         cfg.Cooldown,
	}
}

// WithProxySessionKey attaches one explicit sticky-proxy session key to a request context.
//
// Blank keys are treated as "unset" and leave the context unchanged.
// A nil context falls back to context.Background().
func WithProxySessionKey(ctx context.Context, key string) context.Context {
	return proxyselector.WithSessionKey(ctx, key)
}

func proxySessionKeyFromRequest(req *http.Request) string {
	return proxyselector.SessionKeyFromRequest(req)
}

// NewStaticProxySelector creates a selector that always returns the same proxy.
//
// Empty strings are treated as "no proxy" and return nil, nil.
func NewStaticProxySelector(rawURL string) (ProxySelector, error) {
	return proxyselector.NewStaticSelector(rawURL)
}

// NewRoundRobinProxySelector creates a selector that rotates across proxies.
//
// Empty strings are ignored. If no valid proxies remain, nil, nil is returned.
func NewRoundRobinProxySelector(rawURLs ...string) (ProxySelector, error) {
	return proxyselector.NewRoundRobinSelector(rawURLs...)
}

type healthCheckedProxySelectorInner interface {
	Next(req *http.Request) (*url.URL, error)
	ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	ProxyAvailable(proxyURL *url.URL, now time.Time) bool
	MetricsSnapshot() proxyselector.MetricsSnapshot
}

type healthCheckedProxySelector struct {
	inner healthCheckedProxySelectorInner
}

// NewHealthCheckedRoundRobinProxySelector creates a round-robin proxy pool with failure scoring and cooldown.
func NewHealthCheckedRoundRobinProxySelector(cfg ProxyHealthConfig, rawURLs ...string) (ProxySelector, error) {
	inner, err := proxyselector.NewHealthCheckedRoundRobinSelector(proxyselector.HealthConfig{
		FailureThreshold: cfg.FailureThreshold,
		Cooldown:         cfg.Cooldown,
	}, rawURLs...)
	if err != nil {
		return nil, err
	}
	if inner == nil {
		return nil, nil
	}
	healthChecked, ok := inner.(healthCheckedProxySelectorInner)
	if !ok {
		return nil, fmt.Errorf("health-checked proxy selector missing health capabilities")
	}
	return &healthCheckedProxySelector{inner: healthChecked}, nil
}

func (s *healthCheckedProxySelector) Next(req *http.Request) (*url.URL, error) {
	proxyURL, err := s.inner.Next(req)
	if errors.Is(err, proxyselector.ErrAllCoolingDown) {
		return nil, ErrAllProxiesCoolingDown
	}
	return proxyURL, err
}

func (s *healthCheckedProxySelector) ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error) {
	s.inner.ReportProxyResult(req, proxyURL, statusCode, err)
}

func (s *healthCheckedProxySelector) ProxyAvailable(proxyURL *url.URL, now time.Time) bool {
	return s.inner.ProxyAvailable(proxyURL, now)
}

func (s *healthCheckedProxySelector) MetricsSnapshot() proxyselector.MetricsSnapshot {
	return s.inner.MetricsSnapshot()
}

func (s *healthCheckedProxySelector) ProxyMetricsSnapshot() ProxyMetricsSnapshot {
	return mapProxyMetricsSnapshot(s.inner.MetricsSnapshot())
}

// NewStickyProxySelector wraps a base selector and keeps one proxy choice per explicit session key.
//
// When no session key is attached to the request context, it falls back to the base selector behavior.
func NewStickyProxySelector(base ProxySelector) ProxySelector {
	if base == nil {
		return nil
	}
	sticky := proxyselector.NewStickySelector(base)
	if metrics, ok := base.(ProxyMetricsProvider); ok {
		return &stickyProxySelectorWithMetrics{
			inner:   sticky,
			metrics: metrics,
		}
	}
	return sticky
}

type stickyProxySelectorWithMetrics struct {
	inner   ProxySelector
	metrics ProxyMetricsProvider
}

func (s *stickyProxySelectorWithMetrics) Next(req *http.Request) (*url.URL, error) {
	return s.inner.Next(req)
}

func (s *stickyProxySelectorWithMetrics) ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error) {
	if reporter, ok := s.inner.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	}); ok {
		reporter.ReportProxyResult(req, proxyURL, statusCode, err)
	}
}

func (s *stickyProxySelectorWithMetrics) ProxyMetricsSnapshot() ProxyMetricsSnapshot {
	return s.metrics.ProxyMetricsSnapshot()
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
	mapped := make([]proxyselector.Route, 0, len(routes))
	for _, route := range routes {
		mapped = append(mapped, proxyselector.Route{
			Host:       route.Host,
			PathPrefix: route.PathPrefix,
			ProxyURL:   route.ProxyURL,
		})
	}
	return proxyselector.NewRoutingSelector(mapped...)
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

func mapProxyMetricsSnapshot(src proxyselector.MetricsSnapshot) ProxyMetricsSnapshot {
	out := ProxyMetricsSnapshot{
		GeneratedAt:    src.GeneratedAt,
		TotalProxies:   src.TotalProxies,
		HealthyProxies: src.HealthyProxies,
		CoolingProxies: src.CoolingProxies,
		Proxies:        make([]ProxyEndpointMetrics, 0, len(src.Proxies)),
	}
	for _, proxy := range src.Proxies {
		out.Proxies = append(out.Proxies, ProxyEndpointMetrics{
			ProxyURL:       proxy.ProxyURL,
			FailureScore:   proxy.FailureScore,
			CooldownUntil:  proxy.CooldownUntil,
			SelectionCount: proxy.SelectionCount,
			SuccessCount:   proxy.SuccessCount,
			FailureCount:   proxy.FailureCount,
			CooldownCount:  proxy.CooldownCount,
			LastFailureAt:  proxy.LastFailureAt,
			LastSuccessAt:  proxy.LastSuccessAt,
		})
	}
	return out
}
