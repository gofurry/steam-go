package steam

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	itransport "github.com/GoFurry/steam-go/internal/transport"
)

const defaultProxyClientTimeout = 10 * time.Second

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
