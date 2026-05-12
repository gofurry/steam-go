package transport

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/GoFurry/steam-go/internal/traffic"
	"golang.org/x/time/rate"
)

// ProxySelector chooses a proxy for a request.
type ProxySelector interface {
	Next(req *http.Request) (*url.URL, error)
}

type proxyResultReporter interface {
	ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
}

// RateLimiterConfig defines one token-bucket limiter.
type RateLimiterConfig struct {
	Limit rate.Limit
	Burst int
}

// RequestControlConfig defines one host/session scoped control block.
type RequestControlConfig struct {
	RateLimiter   RateLimiterConfig
	MaxConcurrent int
}

// ClientConfig defines one transport client behavior bundle.
type ClientConfig struct {
	RateLimiter    RateLimiterConfig
	HostControl    RequestControlConfig
	SessionControl RequestControlConfig
}

// Client applies retry and rate limiting on top of an http.Client.
type Client struct {
	httpClient        *http.Client
	limiter           *rate.Limiter
	hostController    *requestControlManager
	sessionController *requestControlManager
}

// New creates a transport client.
func New(httpClient *http.Client, cfg ClientConfig) *Client {
	var limiter *rate.Limiter
	if cfg.RateLimiter.Limit > 0 && cfg.RateLimiter.Burst > 0 {
		limiter = rate.NewLimiter(cfg.RateLimiter.Limit, cfg.RateLimiter.Burst)
	}
	return &Client{
		httpClient:        httpClient,
		limiter:           limiter,
		hostController:    newRequestControlManager(cfg.HostControl),
		sessionController: newRequestControlManager(cfg.SessionControl),
	}
}

// Do executes a single request with optional rate limiting.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if c == nil || c.httpClient == nil {
		return nil, fmt.Errorf("http client is required")
	}
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	hostKey := requestHostKey(req)
	sessionKey := requestSessionKey(ctx)

	if err := waitRequestControl(ctx, c.hostController, hostKey); err != nil {
		return nil, err
	}
	if err := waitRequestControl(ctx, c.sessionController, sessionKey); err != nil {
		return nil, err
	}
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}
	}

	hostRelease, err := acquireRequestControl(ctx, c.hostController, hostKey)
	if err != nil {
		return nil, err
	}
	defer hostRelease()

	sessionRelease, err := acquireRequestControl(ctx, c.sessionController, sessionKey)
	if err != nil {
		return nil, err
	}
	defer sessionRelease()

	return c.httpClient.Do(req.Clone(ctx))
}

// WrapRoundTripper installs proxy selection on top of an existing transport.
func WrapRoundTripper(rt http.RoundTripper, selector ProxySelector) (http.RoundTripper, error) {
	if selector == nil {
		if rt != nil {
			return rt, nil
		}
		return defaultTransport(), nil
	}

	base, err := baseTransport(rt)
	if err != nil {
		return nil, err
	}
	base.Proxy = func(req *http.Request) (*url.URL, error) {
		return selectedProxyFromContext(req.Context()), nil
	}
	return proxySelectionRoundTripper{
		base:     base,
		selector: selector,
	}, nil
}

func baseTransport(rt http.RoundTripper) (*http.Transport, error) {
	if rt == nil {
		return defaultTransport(), nil
	}
	if transport, ok := rt.(*http.Transport); ok {
		return transport.Clone(), nil
	}
	return nil, fmt.Errorf("proxy selector requires an *http.Transport or nil transport, got %T", rt)
}

func defaultTransport() *http.Transport {
	if base, ok := http.DefaultTransport.(*http.Transport); ok {
		return base.Clone()
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

type selectedProxyContextKey struct{}

type proxySelectionRoundTripper struct {
	base     http.RoundTripper
	selector ProxySelector
}

func (rt proxySelectionRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	selectedProxy, err := rt.selector.Next(req)
	if err != nil {
		return nil, err
	}

	reqWithProxy := req.Clone(context.WithValue(req.Context(), selectedProxyContextKey{}, cloneURL(selectedProxy)))
	resp, roundTripErr := rt.base.RoundTrip(reqWithProxy)
	if reporter, ok := rt.selector.(proxyResultReporter); ok {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		reporter.ReportProxyResult(req, cloneURL(selectedProxy), statusCode, roundTripErr)
	}
	return resp, roundTripErr
}

func selectedProxyFromContext(ctx context.Context) *url.URL {
	if ctx == nil {
		return nil
	}
	proxyURL, _ := ctx.Value(selectedProxyContextKey{}).(*url.URL)
	return cloneURL(proxyURL)
}

func cloneURL(proxyURL *url.URL) *url.URL {
	if proxyURL == nil {
		return nil
	}
	cloned := *proxyURL
	return &cloned
}

func requestHostKey(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	return req.URL.Host
}

func requestSessionKey(ctx context.Context) string {
	key, ok := traffic.RequestSessionKeyFromContext(ctx)
	if !ok {
		return ""
	}
	return key
}
