package request

import (
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
	store(req *http.Request, resp *http.Response, body []byte, now time.Time)
	refresh(lookup cacheLookup, resp *http.Response, now time.Time) ([]byte, bool)
}

type cacheLookup struct {
	key          string
	body         []byte
	etag         string
	lastModified string
	fresh        bool
	found        bool
}

type memoryCacheRuntime struct {
	ttl       time.Duration
	cookieJar http.CookieJar

	mu      sync.RWMutex
	entries map[string]cacheEntry

	opCount   atomic.Uint64
	lastSweep atomic.Int64
}

type cacheEntry struct {
	body         []byte
	etag         string
	lastModified string
	storedAt     time.Time
	expiresAt    time.Time
}

const (
	memoryCacheMaxEntries       = 512
	memoryCacheSweepIntervalOps = 64
)

func NewMemoryCacheRuntime(ttl time.Duration, jar http.CookieJar) CacheRuntime {
	if ttl <= 0 {
		return nil
	}
	return &memoryCacheRuntime{
		ttl:       ttl,
		cookieJar: jar,
		entries:   make(map[string]cacheEntry),
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

	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return cacheLookup{}
	}
	if entry.expiresAt.Before(now) && entry.etag == "" && entry.lastModified == "" {
		c.mu.Lock()
		current, ok := c.entries[key]
		if ok && current.expiresAt.Before(now) && current.etag == "" && current.lastModified == "" {
			delete(c.entries, key)
		}
		c.mu.Unlock()
		return cacheLookup{}
	}

	return cacheLookup{
		key:          key,
		body:         cloneBytes(entry.body),
		etag:         entry.etag,
		lastModified: entry.lastModified,
		fresh:        !entry.expiresAt.Before(now),
		found:        true,
	}
}

func (c *memoryCacheRuntime) store(req *http.Request, resp *http.Response, body []byte, now time.Time) {
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
		body:         cloneBytes(body),
		etag:         strings.TrimSpace(resp.Header.Get("ETag")),
		lastModified: strings.TrimSpace(resp.Header.Get("Last-Modified")),
		storedAt:     now,
		expiresAt:    now.Add(c.ttl),
	}
	if len(c.entries) > memoryCacheMaxEntries {
		c.pruneLocked(now, len(c.entries)-memoryCacheMaxEntries)
	}
	c.mu.Unlock()
}

func (c *memoryCacheRuntime) refresh(lookup cacheLookup, resp *http.Response, now time.Time) ([]byte, bool) {
	if c == nil || !lookup.found || lookup.key == "" {
		return nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[lookup.key]
	if !ok {
		return nil, false
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
	return cloneBytes(entry.body), true
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
	if c.cookieJar != nil {
		jarCookies = hashCacheDimension(normalizedCookieKey(c.cookieJar.Cookies(req.URL)))
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
			continue
		}
		candidates = append(candidates, candidate{
			key:       key,
			expires:   entry.expiresAt,
			removable: entry.expiresAt.Before(now),
		})
	}

	if targetExtra <= 0 || len(c.entries) <= memoryCacheMaxEntries {
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
		toRemove--
	}
}
