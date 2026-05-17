package steam_test

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	steam "github.com/gofurry/steam-go"
)

func TestDoRawHTTPRequestAppliesHeaderProfileAndDynamicCookieJar(t *testing.T) {
	t.Parallel()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}

	var sawInitialHeader atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			if got := r.Header.Get("Accept-Language"); got != steam.DefaultPublicStoreHeaderProfileZH().AcceptLanguage {
				t.Fatalf("unexpected accept-language: %q", got)
			}
			sawInitialHeader.Store(true)
			http.SetCookie(w, &http.Cookie{Name: "sessionid", Value: "store-session", Path: "/"})
			http.Redirect(w, r, "/final", http.StatusFound)
		case "/final":
			if _, err := r.Cookie("sessionid"); err != nil {
				t.Fatalf("expected redirected request to carry cookie jar state: %v", err)
			}
			w.Header().Set("ETag", `"etag-a"`)
			_, _ = w.Write([]byte("ok"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	profile := steam.DefaultPublicStoreHeaderProfileZH()
	client, err := steam.NewClient(
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			BlockPolicy:   &steam.TrafficBlockPolicy{},
			HeaderProfile: &profile,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/redirect", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	result, err := client.DoRawHTTPRequest(context.Background(), req, &steam.RawHTTPRequestOptions{
		TrafficClass: steam.TrafficClassPublicStorePage,
		CookieJar:    jar,
	})
	if err != nil {
		t.Fatalf("DoRawHTTPRequest returned error: %v", err)
	}
	if !sawInitialHeader.Load() {
		t.Fatal("expected initial request to receive the class header profile")
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", result.StatusCode)
	}
	if got := string(result.Body); got != "ok" {
		t.Fatalf("unexpected body: %q", got)
	}
	if result.FinalURL == nil || result.FinalURL.Path != "/final" {
		t.Fatalf("unexpected final url: %#v", result.FinalURL)
	}

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}
	cookies := jar.Cookies(parsed)
	if len(cookies) == 0 || cookies[0].Name != "sessionid" || cookies[0].Value != "store-session" {
		t.Fatalf("expected cookie jar to persist redirect cookies, got %#v", cookies)
	}
}

func TestDoRawHTTPRequestRetriesRetryableBlock(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch attempts.Add(1) {
		case 1:
			http.Error(w, "access denied", http.StatusForbidden)
		case 2:
			_, _ = w.Write([]byte("recovered"))
		default:
			t.Fatalf("unexpected extra request")
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithRetry(1),
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			BlockPolicy: &steam.TrafficBlockPolicy{},
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/recover", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	result, err := client.DoRawHTTPRequest(context.Background(), req, &steam.RawHTTPRequestOptions{
		TrafficClass: steam.TrafficClassPublicStorePage,
	})
	if err != nil {
		t.Fatalf("DoRawHTTPRequest returned error: %v", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("expected one retry, got %d attempts", got)
	}
	if result.Block != nil {
		t.Fatalf("expected recovered response without block metadata, got %#v", result.Block)
	}
	if got := string(result.Body); got != "recovered" {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestDoRawHTTPRequestReturnsRawBlockMetadataWithoutError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "access denied", http.StatusForbidden)
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			BlockPolicy: &steam.TrafficBlockPolicy{},
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/blocked", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	result, err := client.DoRawHTTPRequest(context.Background(), req, &steam.RawHTTPRequestOptions{
		TrafficClass: steam.TrafficClassPublicStorePage,
	})
	if err != nil {
		t.Fatalf("DoRawHTTPRequest returned error: %v", err)
	}
	if result.StatusCode != http.StatusForbidden {
		t.Fatalf("unexpected status: %d", result.StatusCode)
	}
	if result.Block == nil {
		t.Fatal("expected block metadata")
	}
	if result.Block.Kind != steam.ErrorKindHTTPStatus || !result.Block.Retryable {
		t.Fatalf("unexpected block metadata: %#v", result.Block)
	}
}

func TestDoRawHTTPRequestCachesFullResultAndRefreshesAfter304(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requestCount.Add(1) {
		case 1:
			w.Header().Set("ETag", `"etag-a"`)
			_, _ = w.Write([]byte("cached"))
		case 2:
			if got := r.Header.Get("If-None-Match"); got != `"etag-a"` {
				t.Fatalf("unexpected If-None-Match: %q", got)
			}
			w.Header().Set("ETag", `"etag-b"`)
			w.WriteHeader(http.StatusNotModified)
		default:
			t.Fatalf("unexpected extra request")
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			Cache:       &steam.TrafficCachePolicy{TTL: 25 * time.Millisecond},
			BlockPolicy: &steam.TrafficBlockPolicy{},
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	makeRequest := func() *http.Request {
		req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/cached", nil)
		if reqErr != nil {
			t.Fatalf("http.NewRequestWithContext returned error: %v", reqErr)
		}
		return req
	}

	first, err := client.DoRawHTTPRequest(context.Background(), makeRequest(), &steam.RawHTTPRequestOptions{
		TrafficClass: steam.TrafficClassPublicStorePage,
	})
	if err != nil {
		t.Fatalf("first DoRawHTTPRequest returned error: %v", err)
	}
	second, err := client.DoRawHTTPRequest(context.Background(), makeRequest(), &steam.RawHTTPRequestOptions{
		TrafficClass: steam.TrafficClassPublicStorePage,
	})
	if err != nil {
		t.Fatalf("second DoRawHTTPRequest returned error: %v", err)
	}
	if got := requestCount.Load(); got != 1 {
		t.Fatalf("expected fresh cache hit, got %d origin requests", got)
	}
	if second.StatusCode != http.StatusOK || string(second.Body) != "cached" {
		t.Fatalf("unexpected cached result: %#v", second)
	}

	time.Sleep(40 * time.Millisecond)

	third, err := client.DoRawHTTPRequest(context.Background(), makeRequest(), &steam.RawHTTPRequestOptions{
		TrafficClass: steam.TrafficClassPublicStorePage,
	})
	if err != nil {
		t.Fatalf("third DoRawHTTPRequest returned error: %v", err)
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("expected one conditional revalidation request, got %d origin requests", got)
	}
	if third.StatusCode != http.StatusOK || string(third.Body) != "cached" {
		t.Fatalf("unexpected refreshed result: %#v", third)
	}
	if third.FinalURL == nil || !strings.HasSuffix(third.FinalURL.Path, "/cached") {
		t.Fatalf("expected cached final url, got %#v", third.FinalURL)
	}
	if got := first.Header.Get("ETag"); got != `"etag-a"` {
		t.Fatalf("unexpected first etag: %q", got)
	}
}

func TestDoRawHTTPRequestValidatesAbsoluteURLAndBodyLimit(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient()
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	relativeReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/relative", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	_, err = client.DoRawHTTPRequest(context.Background(), relativeReq, nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("abcdefghijklmnopqrstuvwxyz"))
	}))
	defer server.Close()

	absoluteReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/body-limit", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	_, err = client.DoRawHTTPRequest(context.Background(), absoluteReq, &steam.RawHTTPRequestOptions{
		MaxResponseBodyBytes: 8,
	})
	expectKind(t, err, steam.ErrorKindTransport)
}
