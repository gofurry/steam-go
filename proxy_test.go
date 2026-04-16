package steam_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
