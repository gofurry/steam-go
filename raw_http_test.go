package steam_test

import (
	"context"
	"io"
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

func TestDoRawHTTPRequestDoesNotRetryPostServerErrorByDefault(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		http.Error(w, "try again", http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := steam.NewClient(steam.WithRetry(1))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/raw-post", strings.NewReader("payload"))
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	result, err := client.DoRawHTTPRequest(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("DoRawHTTPRequest returned error: %v", err)
	}
	if result.StatusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", result.StatusCode)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("expected no POST retry by default, got %d attempts", got)
	}
}

func TestDoRawHTTPRequestRetriesPostServerErrorWhenExplicitlyRetryable(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch attempts.Add(1) {
		case 1:
			http.Error(w, "try again", http.StatusInternalServerError)
		case 2:
			_, _ = w.Write([]byte("ok"))
		default:
			t.Fatalf("unexpected extra request")
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(steam.WithRetry(1))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/raw-post", strings.NewReader("payload"))
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	retryable := true
	result, err := client.DoRawHTTPRequest(context.Background(), req, &steam.RawHTTPRequestOptions{
		Retryable: &retryable,
	})
	if err != nil {
		t.Fatalf("DoRawHTTPRequest returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK || string(result.Body) != "ok" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("expected explicit POST retry, got %d attempts", got)
	}
}

func TestDoRawHTTPRequestUsesGetBodyOnlyPerAttempt(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll returned error: %v", err)
		}
		if got := string(body); got != "payload" {
			t.Fatalf("unexpected body: %q", got)
		}
		switch attempts.Add(1) {
		case 1:
			http.Error(w, "try again", http.StatusInternalServerError)
		case 2:
			_, _ = w.Write([]byte("ok"))
		default:
			t.Fatalf("unexpected extra request")
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(steam.WithRetry(1))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	var getBodyCalls atomic.Int32
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/raw-post", io.NopCloser(strings.NewReader("unused")))
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	req.GetBody = func() (io.ReadCloser, error) {
		getBodyCalls.Add(1)
		return io.NopCloser(strings.NewReader("payload")), nil
	}

	retryable := true
	result, err := client.DoRawHTTPRequest(context.Background(), req, &steam.RawHTTPRequestOptions{
		Retryable: &retryable,
	})
	if err != nil {
		t.Fatalf("DoRawHTTPRequest returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", result.StatusCode)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("expected two attempts, got %d", got)
	}
	if got := getBodyCalls.Load(); got != 2 {
		t.Fatalf("expected GetBody once per attempt, got %d", got)
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

func TestDoRawHTTPRequestHostPolicyAllowsOnlyConfiguredHosts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("allowed"))
	}))
	defer server.Close()

	policy, err := steam.NewAllowedRawHTTPHostPolicy(server.URL)
	if err != nil {
		t.Fatalf("NewAllowedRawHTTPHostPolicy returned error: %v", err)
	}
	client, err := steam.NewClient()
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	allowedReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/allowed", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	result, err := client.DoRawHTTPRequest(context.Background(), allowedReq, &steam.RawHTTPRequestOptions{
		HostPolicy: policy,
	})
	if err != nil {
		t.Fatalf("DoRawHTTPRequest returned error: %v", err)
	}
	if string(result.Body) != "allowed" {
		t.Fatalf("unexpected body: %q", string(result.Body))
	}

	rejectedReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/rejected", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	_, err = client.DoRawHTTPRequest(context.Background(), rejectedReq, &steam.RawHTTPRequestOptions{
		HostPolicy: policy,
	})
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestSuffixRawHTTPHostPolicyMatchesBoundarySafely(t *testing.T) {
	t.Parallel()

	policy, err := steam.NewSuffixRawHTTPHostPolicy("steamstatic.com")
	if err != nil {
		t.Fatalf("NewSuffixRawHTTPHostPolicy returned error: %v", err)
	}

	tests := []struct {
		rawURL string
		allow  bool
	}{
		{rawURL: "https://steamstatic.com/path", allow: true},
		{rawURL: "https://shared.steamstatic.com/path", allow: true},
		{rawURL: "https://shared.steamstatic.com:443/path", allow: true},
		{rawURL: "https://evilsteamstatic.com/path", allow: false},
		{rawURL: "https://steamstatic.com.evil.example/path", allow: false},
	}
	for _, tc := range tests {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tc.rawURL, nil)
		if err != nil {
			t.Fatalf("NewRequestWithContext returned error: %v", err)
		}
		err = policy.Allow(req)
		if tc.allow && err != nil {
			t.Fatalf("expected %s to be allowed: %v", tc.rawURL, err)
		}
		if !tc.allow && err == nil {
			t.Fatalf("expected %s to be rejected", tc.rawURL)
		}
	}
}

func TestNewSuffixRawHTTPHostPolicyRejectsInvalidSuffixes(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"*",
		"*.steamstatic.com",
		"steamstatic.com/path",
		"steamstatic.com?x=1",
		"https://steamstatic.com/path",
		"steamstatic.com:443",
	}
	for _, suffix := range tests {
		if _, err := steam.NewSuffixRawHTTPHostPolicy(suffix); err == nil {
			t.Fatalf("expected suffix %q to be rejected", suffix)
		}
	}
}

func TestSteamStaticRawHTTPHostPolicyAllowsCommonStaticHosts(t *testing.T) {
	t.Parallel()

	policy := steam.NewSteamStaticRawHTTPHostPolicy()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://cdn.cloudflare.steamstatic.com/path", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}
	if err := policy.Allow(req); err != nil {
		t.Fatalf("expected steamstatic host to be allowed: %v", err)
	}
}

func TestNewAllowedRawHTTPHostPolicyRejectsInvalidHosts(t *testing.T) {
	t.Parallel()

	if _, err := steam.NewAllowedRawHTTPHostPolicy(""); err == nil {
		t.Fatal("expected empty host to be rejected")
	}
	if _, err := steam.NewAllowedRawHTTPHostPolicy("example.com/path"); err == nil {
		t.Fatal("expected host with path to be rejected")
	}
}
