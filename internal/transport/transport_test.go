package transport

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GoFurry/steam-go/internal/traffic"
)

func TestWrapRoundTripperSupportsHTTPSProxy(t *testing.T) {
	t.Parallel()

	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`ok`))
	}))
	defer target.Close()

	var proxyCalls atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect {
			t.Fatalf("unexpected proxy method: %s", r.Method)
		}
		proxyCalls.Add(1)

		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("proxy response writer does not support hijacking")
		}
		clientConn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("Hijack returned error: %v", err)
		}

		targetConn, err := net.Dial("tcp", target.Listener.Addr().String())
		if err != nil {
			_ = clientConn.Close()
			t.Fatalf("Dial target returned error: %v", err)
		}

		_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		go proxyCopy(clientConn, targetConn)
		go proxyCopy(targetConn, clientConn)
	}))
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatalf("Parse proxy URL returned error: %v", err)
	}

	base := target.Client().Transport.(*http.Transport)
	rt, err := WrapRoundTripper(base, staticProxySelector{url: proxyURL})
	if err != nil {
		t.Fatalf("WrapRoundTripper returned error: %v", err)
	}
	client := &http.Client{
		Transport: rt,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target.URL, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do returned error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if got, want := string(body), "ok"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
	if proxyCalls.Load() == 0 {
		t.Fatal("expected proxy CONNECT to be used")
	}
}

func TestWrapRoundTripperPreservesCustomRoundTripperWithoutSelector(t *testing.T) {
	t.Parallel()

	custom := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})

	rt, err := WrapRoundTripper(custom, nil)
	if err != nil {
		t.Fatalf("WrapRoundTripper returned error: %v", err)
	}

	resp, err := rt.RoundTrip(httptest.NewRequest(http.MethodGet, "https://example.com", nil))
	if err != nil {
		t.Fatalf("RoundTrip returned error: %v", err)
	}
	if got, want := resp.StatusCode, http.StatusNoContent; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
}

func TestWrapRoundTripperRejectsCustomRoundTripperWithProxySelector(t *testing.T) {
	t.Parallel()

	custom := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("should not be called")
	})

	_, err := WrapRoundTripper(custom, staticProxySelector{url: &url.URL{Scheme: "http", Host: "127.0.0.1:7897"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWrapRoundTripperCallsSelectorOncePerRequest(t *testing.T) {
	t.Parallel()

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`ok`))
	}))
	defer target.Close()

	selector := &reportingProxySelector{}
	base := target.Client().Transport.(*http.Transport)
	rt, err := WrapRoundTripper(base, selector)
	if err != nil {
		t.Fatalf("WrapRoundTripper returned error: %v", err)
	}

	client := &http.Client{Transport: rt}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target.URL, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do returned error: %v", err)
	}
	_ = resp.Body.Close()

	if selector.calls.Load() != 1 {
		t.Fatalf("expected one selector.Next call, got %d", selector.calls.Load())
	}
	if selector.reportCalls.Load() != 1 {
		t.Fatalf("expected one report call, got %d", selector.reportCalls.Load())
	}
	if got := selector.lastStatus.Load(); got != http.StatusOK {
		t.Fatalf("unexpected reported status: %d", got)
	}
}

func TestWrapRoundTripperReportsTransportErrors(t *testing.T) {
	t.Parallel()

	selector := &reportingProxySelector{}
	base := &http.Transport{
		DialContext: func(context.Context, string, string) (net.Conn, error) {
			return nil, errors.New("dial failed")
		},
	}
	rt, err := WrapRoundTripper(base, selector)
	if err != nil {
		t.Fatalf("WrapRoundTripper returned error: %v", err)
	}

	client := &http.Client{Transport: rt}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}

	if _, err := client.Do(req); err == nil {
		t.Fatal("expected transport error")
	}
	if selector.calls.Load() != 1 {
		t.Fatalf("expected one selector.Next call, got %d", selector.calls.Load())
	}
	if selector.reportCalls.Load() != 1 {
		t.Fatalf("expected one report call, got %d", selector.reportCalls.Load())
	}
	if !selector.sawErr.Load() {
		t.Fatal("expected transport error to be reported")
	}
}

func TestWrapRoundTripperReportsHTTPStatusCodes(t *testing.T) {
	t.Parallel()

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer target.Close()

	selector := &reportingProxySelector{}
	base := target.Client().Transport.(*http.Transport)
	rt, err := WrapRoundTripper(base, selector)
	if err != nil {
		t.Fatalf("WrapRoundTripper returned error: %v", err)
	}

	client := &http.Client{Transport: rt}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target.URL, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do returned error: %v", err)
	}
	_ = resp.Body.Close()

	if selector.reportCalls.Load() != 1 {
		t.Fatalf("expected one report call, got %d", selector.reportCalls.Load())
	}
	if got := selector.lastStatus.Load(); got != http.StatusTooManyRequests {
		t.Fatalf("unexpected reported status: %d", got)
	}
	if selector.sawErr.Load() {
		t.Fatal("did not expect transport error for HTTP response")
	}
}

func TestClientHostConcurrencySerializesSameHost(t *testing.T) {
	t.Parallel()

	firstStarted := make(chan struct{})
	firstRelease := make(chan struct{})
	secondStarted := make(chan struct{})

	client := New(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/one":
				close(firstStarted)
				<-firstRelease
			case "/two":
				close(secondStarted)
			}
			return okResponse(), nil
		}),
	}, ClientConfig{
		HostControl: RequestControlConfig{MaxConcurrent: 1},
	})

	firstDone := make(chan error, 1)
	go func() {
		firstDone <- doTransportRequest(client, context.Background(), "https://api.steampowered.com/one")
	}()
	<-firstStarted

	secondDone := make(chan error, 1)
	go func() {
		secondDone <- doTransportRequest(client, context.Background(), "https://api.steampowered.com/two")
	}()

	select {
	case <-secondStarted:
		t.Fatal("expected second request to wait for host concurrency slot")
	case <-time.After(40 * time.Millisecond):
	}

	close(firstRelease)

	select {
	case <-secondStarted:
	case <-time.After(time.Second):
		t.Fatal("expected second request to resume after host slot release")
	}

	if err := <-firstDone; err != nil {
		t.Fatalf("first request returned error: %v", err)
	}
	if err := <-secondDone; err != nil {
		t.Fatalf("second request returned error: %v", err)
	}
}

func TestClientHostConcurrencyIsolatesDifferentHosts(t *testing.T) {
	t.Parallel()

	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	release := make(chan struct{})

	client := New(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Host {
			case "api-a.steampowered.com":
				close(firstStarted)
			case "api-b.steampowered.com":
				close(secondStarted)
			}
			<-release
			return okResponse(), nil
		}),
	}, ClientConfig{
		HostControl: RequestControlConfig{MaxConcurrent: 1},
	})

	firstDone := make(chan error, 1)
	secondDone := make(chan error, 1)
	go func() {
		firstDone <- doTransportRequest(client, context.Background(), "https://api-a.steampowered.com/one")
	}()
	go func() {
		secondDone <- doTransportRequest(client, context.Background(), "https://api-b.steampowered.com/two")
	}()

	select {
	case <-firstStarted:
	case <-time.After(time.Second):
		t.Fatal("expected first host request to start")
	}
	select {
	case <-secondStarted:
	case <-time.After(time.Second):
		t.Fatal("expected second host request to start without waiting")
	}

	close(release)
	if err := <-firstDone; err != nil {
		t.Fatalf("first request returned error: %v", err)
	}
	if err := <-secondDone; err != nil {
		t.Fatalf("second request returned error: %v", err)
	}
}

func TestClientHostRateLimiterThrottlesSameHost(t *testing.T) {
	t.Parallel()

	client := New(&http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return okResponse(), nil
		}),
	}, ClientConfig{
		HostControl: RequestControlConfig{
			RateLimiter: RateLimiterConfig{Limit: 1, Burst: 1},
		},
	})

	if err := doTransportRequest(client, context.Background(), "https://api.steampowered.com/one"); err != nil {
		t.Fatalf("first request returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := doTransportRequest(client, ctx, "https://api.steampowered.com/two")
	if err == nil || !strings.Contains(err.Error(), "would exceed context deadline") {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestClientSessionConcurrencySerializesSameSession(t *testing.T) {
	t.Parallel()

	firstStarted := make(chan struct{})
	firstRelease := make(chan struct{})
	secondStarted := make(chan struct{})

	client := New(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/one":
				close(firstStarted)
				<-firstRelease
			case "/two":
				close(secondStarted)
			}
			return okResponse(), nil
		}),
	}, ClientConfig{
		SessionControl: RequestControlConfig{MaxConcurrent: 1},
	})

	firstCtx := traffic.WithRequestSessionKey(context.Background(), "session-a")
	secondCtx := traffic.WithRequestSessionKey(context.Background(), "session-a")

	firstDone := make(chan error, 1)
	go func() {
		firstDone <- doTransportRequest(client, firstCtx, "https://api.steampowered.com/one")
	}()
	<-firstStarted

	secondDone := make(chan error, 1)
	go func() {
		secondDone <- doTransportRequest(client, secondCtx, "https://api.steampowered.com/two")
	}()

	select {
	case <-secondStarted:
		t.Fatal("expected second request to wait for session concurrency slot")
	case <-time.After(40 * time.Millisecond):
	}

	close(firstRelease)

	select {
	case <-secondStarted:
	case <-time.After(time.Second):
		t.Fatal("expected second request to resume after session slot release")
	}

	if err := <-firstDone; err != nil {
		t.Fatalf("first request returned error: %v", err)
	}
	if err := <-secondDone; err != nil {
		t.Fatalf("second request returned error: %v", err)
	}
}

func TestClientSessionConcurrencyIsolatesDifferentSessions(t *testing.T) {
	t.Parallel()

	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	release := make(chan struct{})

	client := New(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/one":
				close(firstStarted)
			case "/two":
				close(secondStarted)
			}
			<-release
			return okResponse(), nil
		}),
	}, ClientConfig{
		SessionControl: RequestControlConfig{MaxConcurrent: 1},
	})

	firstDone := make(chan error, 1)
	secondDone := make(chan error, 1)
	go func() {
		firstDone <- doTransportRequest(client, traffic.WithRequestSessionKey(context.Background(), "session-a"), "https://api.steampowered.com/one")
	}()
	go func() {
		secondDone <- doTransportRequest(client, traffic.WithRequestSessionKey(context.Background(), "session-b"), "https://api.steampowered.com/two")
	}()

	select {
	case <-firstStarted:
	case <-time.After(time.Second):
		t.Fatal("expected first session request to start")
	}
	select {
	case <-secondStarted:
	case <-time.After(time.Second):
		t.Fatal("expected second session request to start without waiting")
	}

	close(release)
	if err := <-firstDone; err != nil {
		t.Fatalf("first request returned error: %v", err)
	}
	if err := <-secondDone; err != nil {
		t.Fatalf("second request returned error: %v", err)
	}
}

func TestClientSessionControlSkipsRequestsWithoutSessionKey(t *testing.T) {
	t.Parallel()

	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	release := make(chan struct{})

	client := New(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/one":
				close(firstStarted)
			case "/two":
				close(secondStarted)
			}
			<-release
			return okResponse(), nil
		}),
	}, ClientConfig{
		SessionControl: RequestControlConfig{MaxConcurrent: 1},
	})

	firstDone := make(chan error, 1)
	secondDone := make(chan error, 1)
	go func() {
		firstDone <- doTransportRequest(client, context.Background(), "https://api.steampowered.com/one")
	}()
	go func() {
		secondDone <- doTransportRequest(client, context.Background(), "https://api.steampowered.com/two")
	}()

	select {
	case <-firstStarted:
	case <-time.After(time.Second):
		t.Fatal("expected first request to start")
	}
	select {
	case <-secondStarted:
	case <-time.After(time.Second):
		t.Fatal("expected second request without session key to skip session control")
	}

	close(release)
	if err := <-firstDone; err != nil {
		t.Fatalf("first request returned error: %v", err)
	}
	if err := <-secondDone; err != nil {
		t.Fatalf("second request returned error: %v", err)
	}
}

func TestClientSessionRateLimiterThrottlesSameSession(t *testing.T) {
	t.Parallel()

	client := New(&http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return okResponse(), nil
		}),
	}, ClientConfig{
		SessionControl: RequestControlConfig{
			RateLimiter: RateLimiterConfig{Limit: 1, Burst: 1},
		},
	})

	sessionCtx := traffic.WithRequestSessionKey(context.Background(), "session-a")
	if err := doTransportRequest(client, sessionCtx, "https://api.steampowered.com/one"); err != nil {
		t.Fatalf("first request returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(traffic.WithRequestSessionKey(context.Background(), "session-a"), 50*time.Millisecond)
	defer cancel()
	err := doTransportRequest(client, ctx, "https://api.steampowered.com/two")
	if err == nil || !strings.Contains(err.Error(), "would exceed context deadline") {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func proxyCopy(dst net.Conn, src net.Conn) {
	_, _ = io.Copy(dst, src)
	_ = dst.Close()
	_ = src.Close()
}

type staticProxySelector struct {
	url *url.URL
}

func (s staticProxySelector) Next(*http.Request) (*url.URL, error) {
	return s.url, nil
}

type reportingProxySelector struct {
	proxyURL    *url.URL
	calls       atomic.Int32
	reportCalls atomic.Int32
	lastStatus  atomic.Int32
	sawErr      atomic.Bool
}

func (s *reportingProxySelector) Next(*http.Request) (*url.URL, error) {
	s.calls.Add(1)
	if s.proxyURL == nil {
		return nil, nil
	}
	cloned := *s.proxyURL
	return &cloned, nil
}

func (s *reportingProxySelector) ReportProxyResult(_ *http.Request, _ *url.URL, statusCode int, err error) {
	s.reportCalls.Add(1)
	s.lastStatus.Store(int32(statusCode))
	s.sawErr.Store(err != nil)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func doTransportRequest(client *Client, ctx context.Context, rawURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(ctx, req)
	if resp != nil {
		_ = resp.Body.Close()
	}
	return err
}

func okResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("ok")),
		Header:     make(http.Header),
	}
}
