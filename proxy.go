package steam

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	itransport "github.com/GoFurry/steam-go/internal/transport"
)

const defaultProxyClientTimeout = 10 * time.Second

type proxySessionKeyContextKey struct{}

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

// NewStickyProxySelector wraps a base selector and keeps one proxy choice per explicit session key.
//
// When no session key is attached to the request context, it falls back to the base selector behavior.
func NewStickyProxySelector(base ProxySelector) ProxySelector {
	if base == nil {
		return nil
	}
	return &stickyProxySelector{
		base:  base,
		cache: make(map[string]*url.URL),
	}
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

	rt, err := itransport.WrapRoundTripper(http.DefaultTransport, selector)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: rt,
	}, nil
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
	return s.urls[current%uint64(len(s.urls))], nil
}

type stickyProxySelector struct {
	base  ProxySelector
	mu    sync.RWMutex
	cache map[string]*url.URL
}

func (s *stickyProxySelector) Next(req *http.Request) (*url.URL, error) {
	key := proxySessionKeyFromRequest(req)
	if key == "" {
		return s.base.Next(req)
	}

	s.mu.RLock()
	cached, ok := s.cache[key]
	s.mu.RUnlock()
	if ok {
		return cloneProxyURL(cached), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cached, ok = s.cache[key]
	if ok {
		return cloneProxyURL(cached), nil
	}

	proxyURL, err := s.base.Next(req)
	if err != nil {
		return nil, err
	}
	s.cache[key] = cloneProxyURL(proxyURL)
	return cloneProxyURL(proxyURL), nil
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
