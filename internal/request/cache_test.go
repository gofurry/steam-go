package request

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofurry/steam-go/internal/traffic"
)

type cacheTestJar struct {
	cookies []*http.Cookie
}

func (j cacheTestJar) SetCookies(*url.URL, []*http.Cookie) {}

func (j cacheTestJar) Cookies(*url.URL) []*http.Cookie {
	return j.cookies
}

func TestMemoryCacheRuntimeReturnsFreshBody(t *testing.T) {
	t.Parallel()

	runtime := NewMemoryCacheRuntime(time.Minute, nil)
	req := newCacheTestRequest(t, "https://store.steampowered.com/app/10")
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	resp.Header.Set("ETag", `"etag-a"`)

	now := time.Unix(1, 0)
	runtime.store(req, resp, newCacheTestResult("body-a"), now)

	lookup := runtime.lookup(req, now.Add(30*time.Second))
	if !lookup.found || !lookup.fresh {
		t.Fatalf("unexpected lookup: %#v", lookup)
	}
	if got := string(lookup.result.Body); got != "body-a" {
		t.Fatalf("unexpected body: %q", got)
	}

	lookup.result.Body[0] = 'B'
	refetched := runtime.lookup(req, now.Add(30*time.Second))
	if got := string(refetched.result.Body); got != "body-a" {
		t.Fatalf("expected cached body clone, got %q", got)
	}
}

func TestMemoryCacheRuntimeProducesConditionalLookupAfterExpiry(t *testing.T) {
	t.Parallel()

	runtime := NewMemoryCacheRuntime(time.Second, nil)
	req := newCacheTestRequest(t, "https://store.steampowered.com/app/10")
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	resp.Header.Set("ETag", `"etag-a"`)
	resp.Header.Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")

	now := time.Unix(1, 0)
	runtime.store(req, resp, newCacheTestResult("body-a"), now)

	lookup := runtime.lookup(req, now.Add(2*time.Second))
	if !lookup.found || lookup.fresh {
		t.Fatalf("unexpected lookup freshness: %#v", lookup)
	}
	if !cacheLookupAllowsConditionalRequest(lookup) {
		t.Fatal("expected conditional cache lookup to be allowed")
	}

	applyConditionalCacheHeaders(req, lookup)
	if got := req.Header.Get("If-None-Match"); got != `"etag-a"` {
		t.Fatalf("unexpected If-None-Match: %q", got)
	}
	if got := req.Header.Get("If-Modified-Since"); got != "Mon, 02 Jan 2006 15:04:05 GMT" {
		t.Fatalf("unexpected If-Modified-Since: %q", got)
	}
}

func TestMemoryCacheRuntimeRefreshesEntryOnNotModified(t *testing.T) {
	t.Parallel()

	runtime := NewMemoryCacheRuntime(time.Second, nil)
	req := newCacheTestRequest(t, "https://store.steampowered.com/app/10")
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	resp.Header.Set("ETag", `"etag-a"`)

	now := time.Unix(1, 0)
	runtime.store(req, resp, newCacheTestResult("body-a"), now)
	lookup := runtime.lookup(req, now.Add(2*time.Second))
	notModifiedResp := &http.Response{StatusCode: http.StatusNotModified, Header: make(http.Header)}
	notModifiedResp.Header.Set("ETag", `"etag-b"`)
	refreshedResult, ok := runtime.refresh(lookup, notModifiedResp, now.Add(2*time.Second))
	if !ok {
		t.Fatal("expected refresh to succeed")
	}
	if got := string(refreshedResult.Body); got != "body-a" {
		t.Fatalf("unexpected refreshed body: %q", got)
	}

	refetched := runtime.lookup(req, now.Add(2500*time.Millisecond))
	if !refetched.fresh {
		t.Fatalf("expected refreshed entry to be fresh, got %#v", refetched)
	}

	staleAgain := runtime.lookup(req, now.Add(4*time.Second))
	applyConditionalCacheHeaders(req, staleAgain)
	if got := req.Header.Get("If-None-Match"); got != `"etag-b"` {
		t.Fatalf("expected refreshed etag, got %q", got)
	}
}

func TestMemoryCacheRuntimeSeparatesKeysBySessionLanguageAndCookies(t *testing.T) {
	t.Parallel()

	baseURL := "https://store.steampowered.com/app/10?l=schinese"
	now := time.Unix(1, 0)
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}

	runtimeA := NewMemoryCacheRuntime(time.Minute, cacheTestJar{cookies: []*http.Cookie{{Name: "session", Value: "a"}}})
	reqA := newCacheTestRequest(t, baseURL)
	reqA.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	reqA.Header.Set("Cookie", "pref=a")
	reqA = reqA.WithContext(traffic.WithRequestSessionKey(context.Background(), "session-a"))
	runtimeA.store(reqA, resp, newCacheTestResult("body-a"), now)

	reqDifferentSession := newCacheTestRequest(t, baseURL)
	reqDifferentSession.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	reqDifferentSession.Header.Set("Cookie", "pref=a")
	reqDifferentSession = reqDifferentSession.WithContext(traffic.WithRequestSessionKey(context.Background(), "session-b"))
	if lookup := runtimeA.lookup(reqDifferentSession, now); lookup.found {
		t.Fatalf("did not expect cache hit across sessions: %#v", lookup)
	}

	reqDifferentLanguage := newCacheTestRequest(t, baseURL)
	reqDifferentLanguage.Header.Set("Accept-Language", "en-US,en;q=0.9")
	reqDifferentLanguage.Header.Set("Cookie", "pref=a")
	reqDifferentLanguage = reqDifferentLanguage.WithContext(traffic.WithRequestSessionKey(context.Background(), "session-a"))
	if lookup := runtimeA.lookup(reqDifferentLanguage, now); lookup.found {
		t.Fatalf("did not expect cache hit across accept-language: %#v", lookup)
	}

	runtimeB := NewMemoryCacheRuntime(time.Minute, cacheTestJar{cookies: []*http.Cookie{{Name: "session", Value: "b"}}})
	reqSameHeaders := newCacheTestRequest(t, baseURL)
	reqSameHeaders.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	reqSameHeaders.Header.Set("Cookie", "pref=a")
	reqSameHeaders = reqSameHeaders.WithContext(traffic.WithRequestSessionKey(context.Background(), "session-a"))
	if lookup := runtimeB.lookup(reqSameHeaders, now); lookup.found {
		t.Fatalf("did not expect cache hit across cookie jar views: %#v", lookup)
	}
}

func TestMemoryCacheRuntimeDoesNotExposeRawCookieValuesInCacheKey(t *testing.T) {
	t.Parallel()

	runtime := NewMemoryCacheRuntime(time.Minute, cacheTestJar{
		cookies: []*http.Cookie{{Name: "session", Value: "jar-secret"}},
	}).(*memoryCacheRuntime)
	req := newCacheTestRequest(t, "https://steamcommunity.com/inventory/1/730/2")
	req.Header.Set("Cookie", "steamLoginSecure=header-secret")

	key, ok := runtime.cacheKey(req)
	if !ok {
		t.Fatal("expected cache key")
	}
	if strings.Contains(key, "header-secret") {
		t.Fatalf("expected cache key to avoid explicit cookie value, got %q", key)
	}
	if strings.Contains(key, "jar-secret") {
		t.Fatalf("expected cache key to avoid jar cookie value, got %q", key)
	}
	if strings.Contains(key, "steamLoginSecure=header-secret") {
		t.Fatalf("expected cache key to avoid raw cookie header, got %q", key)
	}
}

func TestMemoryCacheRuntimePrunesExpiredEntriesWithoutValidators(t *testing.T) {
	t.Parallel()

	runtime := NewMemoryCacheRuntime(time.Second, nil).(*memoryCacheRuntime)
	now := time.Unix(10, 0)

	reqExpiring := newCacheTestRequest(t, "https://store.steampowered.com/app/10")
	runtime.store(reqExpiring, &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}, newCacheTestResult("body-a"), now)

	reqValidated := newCacheTestRequest(t, "https://store.steampowered.com/app/20")
	respValidated := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	respValidated.Header.Set("ETag", `"etag-a"`)
	runtime.store(reqValidated, respValidated, newCacheTestResult("body-b"), now)

	_ = runtime.lookup(reqExpiring, now.Add(2*time.Second))

	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	if len(runtime.entries) != 1 {
		t.Fatalf("expected one retained cache entry, got %d", len(runtime.entries))
	}
}

func TestMemoryCacheRuntimeCapsEntryCount(t *testing.T) {
	t.Parallel()

	runtime := NewMemoryCacheRuntimeWithOptions(CacheOptions{TTL: time.Minute, MaxEntries: 8}, nil).(*memoryCacheRuntime)
	now := time.Unix(20, 0)
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}

	for i := 0; i < 32; i++ {
		req := newCacheTestRequest(t, "https://store.steampowered.com/app/"+time.Unix(int64(i), 0).Format("150405"))
		runtime.store(req, resp, newCacheTestResult("body"), now.Add(time.Duration(i)*time.Second))
	}

	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	if len(runtime.entries) > 8 {
		t.Fatalf("expected cache size <= 8, got %d", len(runtime.entries))
	}
	if evictions := runtime.Stats().Evictions; evictions == 0 {
		t.Fatal("expected cache evictions to be counted")
	}
}

func TestMemoryCacheRuntimeStatsCountsHitsMissesStoresAndConditionalHits(t *testing.T) {
	t.Parallel()

	runtime := NewMemoryCacheRuntime(time.Second, nil).(*memoryCacheRuntime)
	req := newCacheTestRequest(t, "https://store.steampowered.com/app/10")
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	resp.Header.Set("ETag", `"etag-a"`)
	now := time.Unix(30, 0)

	if lookup := runtime.lookup(req, now); lookup.found {
		t.Fatal("did not expect initial cache hit")
	}
	runtime.store(req, resp, newCacheTestResult("body-a"), now)
	if lookup := runtime.lookup(req, now.Add(500*time.Millisecond)); !lookup.found || !lookup.fresh {
		t.Fatalf("expected fresh cache hit, got %#v", lookup)
	}
	lookup := runtime.lookup(req, now.Add(2*time.Second))
	if !lookup.found || lookup.fresh {
		t.Fatalf("expected stale conditional lookup, got %#v", lookup)
	}
	if _, ok := runtime.refresh(lookup, &http.Response{StatusCode: http.StatusNotModified, Header: make(http.Header)}, now.Add(2*time.Second)); !ok {
		t.Fatal("expected conditional refresh")
	}

	stats := runtime.Stats()
	if stats.Entries != 1 || stats.MaxEntries != defaultMemoryCacheMaxEntries {
		t.Fatalf("unexpected cache size stats: %#v", stats)
	}
	if stats.Misses != 1 || stats.Hits != 1 || stats.Stores != 1 || stats.ConditionalHits != 1 {
		t.Fatalf("unexpected cache counters: %#v", stats)
	}
}

func TestMemoryCacheRuntimeSingleflightSharesSuccessfulFill(t *testing.T) {
	t.Parallel()

	runtime := NewMemoryCacheRuntimeWithOptions(CacheOptions{
		TTL:          time.Minute,
		Singleflight: true,
	}, nil).(*memoryCacheRuntime)
	transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		time.Sleep(20 * time.Millisecond)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("body")),
		}, nil
	})
	policy := ExecutionPolicy{
		RetryBackoff: DefaultRetryBackoffConfig(),
		Transport:    transport,
		CacheRuntime: runtime,
	}

	var calls atomic.Int32
	transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		calls.Add(1)
		time.Sleep(20 * time.Millisecond)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("body")),
			Request:    req,
		}, nil
	})
	policy.Transport = transport

	var wg sync.WaitGroup
	results := make(chan string, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := newCacheTestRequest(t, "https://store.steampowered.com/app/10")
			result, err := ExecuteRawHTTPRequest(context.Background(), req, 1024, policy, nil)
			if err != nil {
				t.Errorf("ExecuteRawHTTPRequest returned error: %v", err)
				return
			}
			results <- string(result.Body)
		}()
	}
	wg.Wait()
	close(results)

	for result := range results {
		if result != "body" {
			t.Fatalf("unexpected result body: %q", result)
		}
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected one upstream call, got %d", got)
	}
}

func TestMemoryCacheRuntimeSingleflightDoesNotCacheFailure(t *testing.T) {
	t.Parallel()

	runtime := NewMemoryCacheRuntimeWithOptions(CacheOptions{
		TTL:          time.Minute,
		Singleflight: true,
	}, nil).(*memoryCacheRuntime)
	var calls atomic.Int32
	policy := ExecutionPolicy{
		RetryBackoff: DefaultRetryBackoffConfig(),
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			if calls.Add(1) == 1 {
				time.Sleep(20 * time.Millisecond)
				return nil, errors.New("temporary failure")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("body")),
			}, nil
		}),
		CacheRuntime: runtime,
	}

	firstDone := make(chan error, 1)
	go func() {
		req := newCacheTestRequest(t, "https://store.steampowered.com/app/20")
		_, err := ExecuteRawHTTPRequest(context.Background(), req, 1024, policy, nil)
		firstDone <- err
	}()
	time.Sleep(5 * time.Millisecond)

	req := newCacheTestRequest(t, "https://store.steampowered.com/app/20")
	result, err := ExecuteRawHTTPRequest(context.Background(), req, 1024, policy, nil)
	if err != nil {
		t.Fatalf("second ExecuteRawHTTPRequest returned error: %v", err)
	}
	if got := string(result.Body); got != "body" {
		t.Fatalf("unexpected second body: %q", got)
	}
	if err := <-firstDone; err == nil {
		t.Fatal("expected first leader request to fail")
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected second request to retry upstream after failed fill, got %d calls", got)
	}
}

func TestMemoryCacheRuntimeSingleflightFollowerRespectsContextCancellation(t *testing.T) {
	t.Parallel()

	runtime := NewMemoryCacheRuntimeWithOptions(CacheOptions{
		TTL:          time.Minute,
		Singleflight: true,
	}, nil).(*memoryCacheRuntime)
	release := make(chan struct{})
	policy := ExecutionPolicy{
		RetryBackoff: DefaultRetryBackoffConfig(),
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			<-release
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("body")),
				Request:    req,
			}, nil
		}),
		CacheRuntime: runtime,
	}

	firstDone := make(chan error, 1)
	go func() {
		req := newCacheTestRequest(t, "https://store.steampowered.com/app/30")
		_, err := ExecuteRawHTTPRequest(context.Background(), req, 1024, policy, nil)
		firstDone <- err
	}()
	time.Sleep(5 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := newCacheTestRequest(t, "https://store.steampowered.com/app/30")
	_, err := ExecuteRawHTTPRequest(ctx, req, 1024, policy, nil)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled follower error, got %v", err)
	}

	close(release)
	if err := <-firstDone; err != nil {
		t.Fatalf("leader returned error: %v", err)
	}
}

func newCacheTestRequest(t *testing.T, rawURL string) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	return req
}

func newCacheTestResult(body string) HTTPResult {
	return HTTPResult{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       []byte(body),
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	return f(req)
}
