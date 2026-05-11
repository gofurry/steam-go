package steam_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	steam "github.com/GoFurry/steam-go"
)

func TestNewStaticProxySelector(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewStaticProxySelector("http://127.0.0.1:7897")
	if err != nil {
		t.Fatalf("NewStaticProxySelector returned error: %v", err)
	}
	if selector == nil {
		t.Fatal("expected selector")
	}

	proxyURL, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if got := proxyURL.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected proxy url: %s", got)
	}
}

func TestNewStaticProxySelectorEmptyReturnsNil(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewStaticProxySelector("  ")
	if err != nil {
		t.Fatalf("NewStaticProxySelector returned error: %v", err)
	}
	if selector != nil {
		t.Fatal("expected nil selector")
	}
}

func TestNewRoundRobinProxySelector(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewRoundRobinProxySelector(
		"http://127.0.0.1:7897",
		"http://127.0.0.1:7898",
	)
	if err != nil {
		t.Fatalf("NewRoundRobinProxySelector returned error: %v", err)
	}
	if selector == nil {
		t.Fatal("expected selector")
	}

	first, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("first Next returned error: %v", err)
	}
	second, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("second Next returned error: %v", err)
	}
	third, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("third Next returned error: %v", err)
	}

	if got := first.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected first proxy: %s", got)
	}
	if got := second.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected second proxy: %s", got)
	}
	if got := third.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected third proxy: %s", got)
	}
}

func TestNewRoundRobinProxySelectorIgnoresEmptyValues(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewRoundRobinProxySelector("", "http://127.0.0.1:7897", " ")
	if err != nil {
		t.Fatalf("NewRoundRobinProxySelector returned error: %v", err)
	}
	if selector == nil {
		t.Fatal("expected selector")
	}

	proxyURL, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if got := proxyURL.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected proxy url: %s", got)
	}
}

func TestNewRoundRobinProxySelectorRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	if _, err := steam.NewRoundRobinProxySelector("://bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestWithProxySessionKeyStoresTrimmedKey(t *testing.T) {
	t.Parallel()

	ctx := steam.WithProxySessionKey(context.Background(), "  session-a  ")
	req := mustRequestWithContext(t, ctx, "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
		},
	}
	selector := steam.NewStickyProxySelector(base)

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if base.calls.Load() != 1 {
		t.Fatalf("expected one base call, got %d", base.calls.Load())
	}
	if got := first.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected first proxy: %s", got)
	}
	if got := second.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected second proxy: %s", got)
	}
}

func TestNewStickyProxySelectorNilBaseReturnsNil(t *testing.T) {
	t.Parallel()

	if selector := steam.NewStickyProxySelector(nil); selector != nil {
		t.Fatal("expected nil selector")
	}
}

func TestStickyProxySelectorPinsPerSessionKey(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
			mustParseURL(t, "http://127.0.0.1:7898"),
		},
	}
	selector := steam.NewStickyProxySelector(base)

	reqA := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")
	reqB := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-b"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	firstA, err := selector.Next(reqA)
	if err != nil {
		t.Fatalf("selector.Next(reqA) returned error: %v", err)
	}
	secondA, err := selector.Next(reqA)
	if err != nil {
		t.Fatalf("selector.Next(reqA) returned error: %v", err)
	}
	firstB, err := selector.Next(reqB)
	if err != nil {
		t.Fatalf("selector.Next(reqB) returned error: %v", err)
	}
	secondB, err := selector.Next(reqB)
	if err != nil {
		t.Fatalf("selector.Next(reqB) returned error: %v", err)
	}

	if got := firstA.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected session-a proxy: %s", got)
	}
	if got := secondA.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected repeated session-a proxy: %s", got)
	}
	if got := firstB.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected session-b proxy: %s", got)
	}
	if got := secondB.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected repeated session-b proxy: %s", got)
	}
	if base.calls.Load() != 2 {
		t.Fatalf("expected two base calls, got %d", base.calls.Load())
	}
}

func TestStickyProxySelectorFallsBackWithoutSessionKey(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
			mustParseURL(t, "http://127.0.0.1:7898"),
		},
	}
	selector := steam.NewStickyProxySelector(base)
	req := mustRequest(t, "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if got := first.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected first proxy: %s", got)
	}
	if got := second.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected second proxy: %s", got)
	}
	if base.calls.Load() != 2 {
		t.Fatalf("expected two base calls, got %d", base.calls.Load())
	}
}

func TestStickyProxySelectorTreatsBlankKeyAsUnset(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
			mustParseURL(t, "http://127.0.0.1:7898"),
		},
	}
	selector := steam.NewStickyProxySelector(base)
	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "   "), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if got := first.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected first proxy: %s", got)
	}
	if got := second.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected second proxy: %s", got)
	}
}

func TestStickyProxySelectorDoesNotCacheErrors(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
		},
		nextErrors: []error{
			errors.New("boom"),
		},
	}
	selector := steam.NewStickyProxySelector(base)
	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	if _, err := selector.Next(req); err == nil {
		t.Fatal("expected error")
	}
	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if got := second.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected recovered proxy: %s", got)
	}
	if base.calls.Load() != 2 {
		t.Fatalf("expected two base calls, got %d", base.calls.Load())
	}
}

func TestStickyProxySelectorCachesDirectConnections(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{nil},
	}
	selector := steam.NewStickyProxySelector(base)
	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if first != nil || second != nil {
		t.Fatal("expected direct connection to remain nil")
	}
	if base.calls.Load() != 1 {
		t.Fatalf("expected one base call, got %d", base.calls.Load())
	}
}

func TestStickyProxySelectorClonesReturnedURLs(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
		},
	}
	selector := steam.NewStickyProxySelector(base)
	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	first.Host = "mutated:1234"

	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if got := second.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected cached proxy after mutation: %s", got)
	}
}

func TestStickyProxySelectorConcurrentSameSessionKey(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
		},
	}
	selector := steam.NewStickyProxySelector(base)

	var wg sync.WaitGroup
	results := make(chan string, 16)
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")
			proxyURL, err := selector.Next(req)
			if err != nil {
				t.Errorf("selector.Next returned error: %v", err)
				return
			}
			if proxyURL == nil {
				t.Error("expected proxy url")
				return
			}
			results <- proxyURL.String()
		}()
	}
	wg.Wait()
	close(results)

	for result := range results {
		if result != "http://127.0.0.1:7897" {
			t.Fatalf("unexpected concurrent proxy: %s", result)
		}
	}
	if base.calls.Load() != 1 {
		t.Fatalf("expected one base call, got %d", base.calls.Load())
	}
}

func TestNewRoutingProxySelector(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewRoutingProxySelector(
		steam.ProxyRoute{
			Host:       "api.steampowered.com",
			PathPrefix: "/ISteamUser/",
			ProxyURL:   "http://127.0.0.1:7897",
		},
		steam.ProxyRoute{
			Host:       "steamcommunity.com",
			PathPrefix: "/openid/",
			ProxyURL:   "",
		},
	)
	if err != nil {
		t.Fatalf("NewRoutingProxySelector returned error: %v", err)
	}
	if selector == nil {
		t.Fatal("expected selector")
	}

	reqAPI := mustRequest(t, "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")
	reqOpenID := mustRequest(t, "https://steamcommunity.com/openid/login")
	reqOther := mustRequest(t, "https://store.steampowered.com/")

	apiProxy, err := selector.Next(reqAPI)
	if err != nil {
		t.Fatalf("selector.Next(api) returned error: %v", err)
	}
	if got := apiProxy.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected api proxy: %s", got)
	}

	openIDProxy, err := selector.Next(reqOpenID)
	if err != nil {
		t.Fatalf("selector.Next(openid) returned error: %v", err)
	}
	if openIDProxy != nil {
		t.Fatal("expected direct route for openid")
	}

	otherProxy, err := selector.Next(reqOther)
	if err != nil {
		t.Fatalf("selector.Next(other) returned error: %v", err)
	}
	if otherProxy != nil {
		t.Fatal("expected direct fallback for unmatched route")
	}
}

func TestNewRoutingProxySelectorRejectsInvalidProxyURL(t *testing.T) {
	t.Parallel()

	_, err := steam.NewRoutingProxySelector(steam.ProxyRoute{
		Host:     "api.steampowered.com",
		ProxyURL: "://bad",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewHTTPClientWithProxySelector(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	selector := &countingProxySelector{}
	httpClient, err := steam.NewHTTPClientWithProxySelector(selector, 2*time.Second)
	if err != nil {
		t.Fatalf("NewHTTPClientWithProxySelector returned error: %v", err)
	}

	resp, err := httpClient.Get(server.URL)
	if err != nil {
		t.Fatalf("httpClient.Get returned error: %v", err)
	}
	_ = resp.Body.Close()

	if got := selector.calls.Load(); got == 0 {
		t.Fatal("expected proxy selector to be used by http client")
	}
}

func TestNewHTTPClientWithProxySelectorRejectsNegativeTimeout(t *testing.T) {
	t.Parallel()

	if _, err := steam.NewHTTPClientWithProxySelector(nil, -time.Second); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewHTTPClientWithProxySelectorDefaultsZeroTimeout(t *testing.T) {
	t.Parallel()

	client, err := steam.NewHTTPClientWithProxySelector(nil, 0)
	if err != nil {
		t.Fatalf("NewHTTPClientWithProxySelector returned error: %v", err)
	}
	if got := client.Timeout; got != 10*time.Second {
		t.Fatalf("unexpected timeout: %s", got)
	}
}

type countingProxySelector struct {
	calls atomic.Int32
}

func (s *countingProxySelector) Next(*http.Request) (*url.URL, error) {
	s.calls.Add(1)
	return nil, nil
}

func mustRequest(t *testing.T, rawURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("http.NewRequest returned error: %v", err)
	}
	return req
}

func mustRequestWithContext(t *testing.T, ctx context.Context, rawURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	return req
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}
	return parsed
}

type recordingProxySelector struct {
	calls       atomic.Int32
	mu          sync.Mutex
	nextResults []*url.URL
	nextErrors  []error
}

func (s *recordingProxySelector) Next(*http.Request) (*url.URL, error) {
	s.calls.Add(1)

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.nextErrors) > 0 {
		err := s.nextErrors[0]
		s.nextErrors = s.nextErrors[1:]
		return nil, err
	}
	if len(s.nextResults) == 0 {
		return nil, nil
	}
	result := s.nextResults[0]
	s.nextResults = s.nextResults[1:]
	if result == nil {
		return nil, nil
	}
	cloned := *result
	return &cloned, nil
}
