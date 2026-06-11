package request

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofurry/steam-go/internal/traffic"
)

type CacheRuntime interface {
	lookup(req *http.Request, now time.Time) cacheLookup
	store(req *http.Request, resp *http.Response, result HTTPResult, now time.Time)
	refresh(lookup cacheLookup, resp *http.Response, now time.Time) (HTTPResult, bool)
	beginFill(ctx context.Context, lookup cacheLookup) (*cacheFillCall, bool)
	completeFill(lookup cacheLookup, call *cacheFillCall, result HTTPResult, stored bool, err error)
	Stats() CacheStats
}

// CacheOptions configures the in-memory conditional-request cache.
type CacheOptions struct {
	TTL          time.Duration
	MaxEntries   int
	Singleflight bool
}

// CacheStats is a read-only snapshot of one cache runtime.
type CacheStats struct {
	Entries         int
	MaxEntries      int
	Hits            uint64
	Misses          uint64
	Stores          uint64
	Evictions       uint64
	ConditionalHits uint64
}

type cacheLookup struct {
	key          string
	result       HTTPResult
	etag         string
	lastModified string
	fresh        bool
	found        bool
	cacheable    bool
}

type memoryCacheRuntime struct {
	ttl          time.Duration
	maxEntries   int
	singleflight bool
	cookieJar    http.CookieJar

	mu      sync.RWMutex
	entries map[string]cacheEntry

	fillMu   sync.Mutex
	inflight map[string]*cacheFillCall

	opCount   atomic.Uint64
	lastSweep atomic.Int64

	hits            atomic.Uint64
	misses          atomic.Uint64
	stores          atomic.Uint64
	evictions       atomic.Uint64
	conditionalHits atomic.Uint64
}

type cacheEntry struct {
	result       HTTPResult
	etag         string
	lastModified string
	storedAt     time.Time
	expiresAt    time.Time
}

const (
	defaultMemoryCacheMaxEntries = 512
	memoryCacheSweepIntervalOps  = 64
)

func NewMemoryCacheRuntime(ttl time.Duration, jar http.CookieJar) CacheRuntime {
	return NewMemoryCacheRuntimeWithOptions(CacheOptions{TTL: ttl}, jar)
}

func NewMemoryCacheRuntimeWithOptions(opts CacheOptions, jar http.CookieJar) CacheRuntime {
	if opts.TTL <= 0 {
		return nil
	}
	maxEntries := opts.MaxEntries
	if maxEntries <= 0 {
		maxEntries = defaultMemoryCacheMaxEntries
	}
	return &memoryCacheRuntime{
		ttl:          opts.TTL,
		maxEntries:   maxEntries,
		singleflight: opts.Singleflight,
		cookieJar:    jar,
		entries:      make(map[string]cacheEntry),
	}
}

func (c *memoryCacheRuntime) lookup(req *http.Request, now time.Time) cacheLookup {
	if c == nil {
		return cacheLookup{}
	}
	c.maybeSweep(now)
	key, ok := c.cacheKey(req)
	if !ok {
		return cacheLookup{}
	}
	lookup := cacheLookup{key: key, cacheable: true}

	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		c.misses.Add(1)
		return lookup
	}
	if entry.expiresAt.Before(now) && entry.etag == "" && entry.lastModified == "" {
		evicted := false
		c.mu.Lock()
		current, ok := c.entries[key]
		if ok && current.expiresAt.Before(now) && current.etag == "" && current.lastModified == "" {
			delete(c.entries, key)
			evicted = true
		}
		c.mu.Unlock()
		if evicted {
			c.evictions.Add(1)
		}
		c.misses.Add(1)
		return lookup
	}

	lookup.result = cloneHTTPResult(entry.result)
	lookup.etag = entry.etag
	lookup.lastModified = entry.lastModified
	lookup.fresh = !entry.expiresAt.Before(now)
	lookup.found = true
	if lookup.fresh {
		c.hits.Add(1)
	}
	return lookup
}

func (c *memoryCacheRuntime) store(req *http.Request, resp *http.Response, result HTTPResult, now time.Time) {
	if c == nil || req == nil || resp == nil || req.Method != http.MethodGet {
		return
	}
	c.maybeSweep(now)
	key, ok := c.cacheKey(req)
	if !ok {
		return
	}

	c.mu.Lock()
	c.entries[key] = cacheEntry{
		result:       cloneHTTPResult(result),
		etag:         strings.TrimSpace(resp.Header.Get("ETag")),
		lastModified: strings.TrimSpace(resp.Header.Get("Last-Modified")),
		storedAt:     now,
		expiresAt:    now.Add(c.ttl),
	}
	c.stores.Add(1)
	if len(c.entries) > c.maxEntries {
		c.pruneLocked(now, len(c.entries)-c.maxEntries)
	}
	c.mu.Unlock()
}

func (c *memoryCacheRuntime) refresh(lookup cacheLookup, resp *http.Response, now time.Time) (HTTPResult, bool) {
	if c == nil || !lookup.found || lookup.key == "" {
		return HTTPResult{}, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[lookup.key]
	if !ok {
		return HTTPResult{}, false
	}

	entry.storedAt = now
	entry.expiresAt = now.Add(c.ttl)
	if resp != nil {
		if etag := strings.TrimSpace(resp.Header.Get("ETag")); etag != "" {
			entry.etag = etag
		}
		if lastModified := strings.TrimSpace(resp.Header.Get("Last-Modified")); lastModified != "" {
			entry.lastModified = lastModified
		}
	}
	c.entries[lookup.key] = entry
	c.conditionalHits.Add(1)
	return cloneHTTPResult(entry.result), true
}

type cacheFillCall struct {
	done   chan struct{}
	result HTTPResult
	stored bool
	err    error
}

func (c *memoryCacheRuntime) beginFill(ctx context.Context, lookup cacheLookup) (*cacheFillCall, bool) {
	if c == nil || !c.singleflight || !lookup.cacheable || lookup.found || lookup.key == "" {
		return nil, true
	}
	if ctx != nil {
		select {
		case <-ctx.Done():
			call := &cacheFillCall{
				done: make(chan struct{}),
				err:  ctx.Err(),
			}
			close(call.done)
			return call, false
		default:
		}
	}

	c.fillMu.Lock()
	defer c.fillMu.Unlock()

	if c.inflight == nil {
		c.inflight = make(map[string]*cacheFillCall)
	}
	if call, ok := c.inflight[lookup.key]; ok {
		return call, false
	}
	call := &cacheFillCall{done: make(chan struct{})}
	c.inflight[lookup.key] = call
	return call, true
}

func (c *memoryCacheRuntime) completeFill(lookup cacheLookup, call *cacheFillCall, result HTTPResult, stored bool, err error) {
	if c == nil || call == nil || lookup.key == "" {
		return
	}
	c.fillMu.Lock()
	if current, ok := c.inflight[lookup.key]; ok && current == call {
		delete(c.inflight, lookup.key)
	}
	call.result = cloneHTTPResult(result)
	call.stored = stored
	call.err = err
	close(call.done)
	c.fillMu.Unlock()
}

func (c *memoryCacheRuntime) Stats() CacheStats {
	if c == nil {
		return CacheStats{}
	}
	c.mu.RLock()
	entries := len(c.entries)
	c.mu.RUnlock()
	return CacheStats{
		Entries:         entries,
		MaxEntries:      c.maxEntries,
		Hits:            c.hits.Load(),
		Misses:          c.misses.Load(),
		Stores:          c.stores.Load(),
		Evictions:       c.evictions.Load(),
		ConditionalHits: c.conditionalHits.Load(),
	}
}

func (call *cacheFillCall) wait(ctx context.Context) (HTTPResult, bool, error) {
	if call == nil {
		return HTTPResult{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-call.done:
		return cloneHTTPResult(call.result), call.stored, call.err
	case <-ctx.Done():
		return HTTPResult{}, false, ctx.Err()
	}
}

func (c *memoryCacheRuntime) cacheKey(req *http.Request) (string, bool) {
	if req == nil || req.URL == nil || req.Method != http.MethodGet {
		return "", false
	}

	var sessionKey string
	if value, ok := traffic.RequestSessionKeyFromContext(req.Context()); ok {
		sessionKey = value
	}
	var jarCookies string
	if jar := c.resolveCookieJar(req.Context()); jar != nil {
		jarCookies = hashCacheDimension(normalizedCookieKey(jar.Cookies(req.URL)))
	}

	acceptLanguage := req.Header.Get("Accept-Language")
	explicitCookie := hashCacheDimension(req.Header.Get("Cookie"))
	rawURL := req.URL.String()

	var builder strings.Builder
	builder.Grow(len(req.Method) + len(rawURL) + len(sessionKey) + len(acceptLanguage) + len(explicitCookie) + len(jarCookies) + 8)
	builder.WriteString(req.Method)
	builder.WriteByte('\x1f')
	builder.WriteString(rawURL)
	builder.WriteByte('\x1f')
	builder.WriteString(sessionKey)
	builder.WriteByte('\x1f')
	builder.WriteString(acceptLanguage)
	builder.WriteByte('\x1f')
	builder.WriteString(explicitCookie)
	builder.WriteByte('\x1f')
	builder.WriteString(jarCookies)
	return builder.String(), true
}

func (c *memoryCacheRuntime) resolveCookieJar(ctx context.Context) http.CookieJar {
	if jar, ok := RuntimeCookieJarFromContext(ctx); ok {
		return jar
	}
	return c.cookieJar
}

func normalizedCookieKey(cookies []*http.Cookie) string {
	if len(cookies) == 0 {
		return ""
	}

	normalized := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		normalized = append(normalized, cookie.Name+"="+cookie.Value)
	}
	sort.Strings(normalized)
	if len(normalized) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, value := range normalized {
		if i > 0 {
			builder.WriteByte('\x1e')
		}
		builder.WriteString(value)
	}
	return builder.String()
}

func hashCacheDimension(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	cloned := make([]byte, len(src))
	copy(cloned, src)
	return cloned
}

func applyConditionalCacheHeaders(req *http.Request, lookup cacheLookup) {
	if req == nil || !lookup.found || lookup.fresh {
		return
	}
	if req.Header.Get("If-None-Match") == "" && lookup.etag != "" {
		req.Header.Set("If-None-Match", lookup.etag)
	}
	if req.Header.Get("If-Modified-Since") == "" && lookup.lastModified != "" {
		req.Header.Set("If-Modified-Since", lookup.lastModified)
	}
}

func cacheLookupAllowsConditionalRequest(lookup cacheLookup) bool {
	return lookup.found && !lookup.fresh && (lookup.etag != "" || lookup.lastModified != "")
}

func requestCacheable(req *http.Request) bool {
	return req != nil && req.Method == http.MethodGet && req.URL != nil
}

func (c *memoryCacheRuntime) maybeSweep(now time.Time) {
	if c == nil {
		return
	}
	count := c.opCount.Add(1)
	if count%memoryCacheSweepIntervalOps != 0 {
		return
	}
	last := c.lastSweep.Load()
	if last != 0 && now.UnixNano()-last < int64(time.Second) {
		return
	}

	c.mu.Lock()
	c.pruneLocked(now, 0)
	c.mu.Unlock()
	c.lastSweep.Store(now.UnixNano())
}

func (c *memoryCacheRuntime) pruneLocked(now time.Time, targetExtra int) {
	if len(c.entries) == 0 {
		return
	}

	type candidate struct {
		key       string
		expires   time.Time
		removable bool
	}

	candidates := make([]candidate, 0, len(c.entries))
	for key, entry := range c.entries {
		expiredNoValidators := entry.expiresAt.Before(now) && entry.etag == "" && entry.lastModified == ""
		if expiredNoValidators {
			delete(c.entries, key)
			c.evictions.Add(1)
			continue
		}
		candidates = append(candidates, candidate{
			key:       key,
			expires:   entry.expiresAt,
			removable: entry.expiresAt.Before(now),
		})
	}

	if targetExtra <= 0 || len(c.entries) <= c.maxEntries {
		return
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].removable != candidates[j].removable {
			return candidates[i].removable
		}
		return candidates[i].expires.Before(candidates[j].expires)
	})

	toRemove := targetExtra
	for _, candidate := range candidates {
		if toRemove <= 0 {
			break
		}
		delete(c.entries, candidate.key)
		c.evictions.Add(1)
		toRemove--
	}
}
