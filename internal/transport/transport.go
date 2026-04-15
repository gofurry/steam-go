package transport

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

// ProxySelector chooses a proxy for a request.
type ProxySelector interface {
	Next(req *http.Request) (*url.URL, error)
}

// Client applies retry and rate limiting on top of an http.Client.
type Client struct {
	httpClient *http.Client
	limiter    *rate.Limiter
}

// New creates a transport client.
func New(httpClient *http.Client, requestsPerSecond int) *Client {
	var limiter *rate.Limiter
	if requestsPerSecond > 0 {
		limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond)
	}
	return &Client{
		httpClient: httpClient,
		limiter:    limiter,
	}
}

// Do executes a single request with optional rate limiting.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}
	}
	return c.httpClient.Do(req.Clone(ctx))
}

// WrapRoundTripper installs proxy selection on top of an existing transport.
func WrapRoundTripper(rt http.RoundTripper, selector ProxySelector) http.RoundTripper {
	base := baseTransport(rt)
	if selector == nil {
		return base
	}
	return &proxyRoundTripper{
		base:     base,
		selector: selector,
	}
}

func baseTransport(rt http.RoundTripper) *http.Transport {
	if rt == nil {
		return defaultTransport()
	}
	if transport, ok := rt.(*http.Transport); ok {
		return transport.Clone()
	}
	return defaultTransport()
}

func defaultTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

type proxyRoundTripper struct {
	base     *http.Transport
	selector ProxySelector
}

func (p *proxyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if p.selector == nil {
		return p.base.RoundTrip(req)
	}

	proxyURL, err := p.selector.Next(req)
	if err != nil {
		return nil, err
	}

	// A fresh struct copy is used here instead of (*http.Transport).Clone().
	// In practice some local HTTP proxies are compatible with a directly
	// configured transport but break when a cloned transport has its Proxy
	// field swapped per request.
	transport := cloneTransport(p.base)
	if proxyURL == nil {
		transport.Proxy = nil
	} else {
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return transport.RoundTrip(req)
}

func cloneTransport(base *http.Transport) *http.Transport {
	if base == nil {
		return defaultTransport()
	}

	cloned := *base
	if base.TLSClientConfig != nil {
		cloned.TLSClientConfig = base.TLSClientConfig.Clone()
	}
	if base.ProxyConnectHeader != nil {
		cloned.ProxyConnectHeader = base.ProxyConnectHeader.Clone()
	}
	if base.TLSNextProto != nil {
		nextProto := make(map[string]func(string, *tls.Conn) http.RoundTripper, len(base.TLSNextProto))
		for k, v := range base.TLSNextProto {
			nextProto[k] = v
		}
		cloned.TLSNextProto = nextProto
	}
	return &cloned
}
