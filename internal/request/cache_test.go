package request

import (
	"context"
	"net/http"
	"net/url"
	"strings"
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

	runtime := NewMemoryCacheRuntime(time.Minute, nil).(*memoryCacheRuntime)
	now := time.Unix(20, 0)
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}

	for i := 0; i < memoryCacheMaxEntries+32; i++ {
		req := newCacheTestRequest(t, "https://store.steampowered.com/app/"+time.Unix(int64(i), 0).Format("150405"))
		runtime.store(req, resp, newCacheTestResult("body"), now.Add(time.Duration(i)*time.Second))
	}

	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	if len(runtime.entries) > memoryCacheMaxEntries {
		t.Fatalf("expected cache size <= %d, got %d", memoryCacheMaxEntries, len(runtime.entries))
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
