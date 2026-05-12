package steam

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/traffic"
	"github.com/GoFurry/steam-go/internal/transport"
	"golang.org/x/time/rate"
)

type stubCookieJar struct{}

func (stubCookieJar) SetCookies(*url.URL, []*http.Cookie) {}

func (stubCookieJar) Cookies(*url.URL) []*http.Cookie { return nil }

func TestWithRateLimitSetsRateLimiterConfig(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithRateLimit(3)(&cfg); err != nil {
		t.Fatalf("WithRateLimit returned error: %v", err)
	}

	want := transport.RateLimiterConfig{
		Limit: rate.Limit(3),
		Burst: 3,
	}
	if cfg.rateLimiter != want {
		t.Fatalf("rateLimiter = %#v, want %#v", cfg.rateLimiter, want)
	}
}

func TestWithRateLimiterDisablesLimiterWithZeroValues(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	cfg.rateLimiter = transport.RateLimiterConfig{
		Limit: rate.Limit(5),
		Burst: 5,
	}
	if err := WithRateLimiter(0, 0)(&cfg); err != nil {
		t.Fatalf("WithRateLimiter returned error: %v", err)
	}
	if cfg.rateLimiter != (transport.RateLimiterConfig{}) {
		t.Fatalf("expected limiter to be disabled, got %#v", cfg.rateLimiter)
	}
}

func TestWithRateLimiterRejectsPartialZeroValues(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithRateLimiter(rate.Limit(1), 0)(&cfg); err == nil {
		t.Fatal("expected error for partial zero limiter config")
	}
	if err := WithRateLimiter(0, 1)(&cfg); err == nil {
		t.Fatal("expected error for partial zero limiter config")
	}
}

func TestLastRateLimiterOptionWins(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithRateLimit(5)(&cfg); err != nil {
		t.Fatalf("WithRateLimit returned error: %v", err)
	}
	if err := WithRateLimiter(rate.Limit(2), 7)(&cfg); err != nil {
		t.Fatalf("WithRateLimiter returned error: %v", err)
	}

	want := transport.RateLimiterConfig{
		Limit: rate.Limit(2),
		Burst: 7,
	}
	if cfg.rateLimiter != want {
		t.Fatalf("rateLimiter = %#v, want %#v", cfg.rateLimiter, want)
	}
}

func TestWithRateLimitCanDisableExplicitLimiter(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithRateLimiter(rate.Limit(2), 7)(&cfg); err != nil {
		t.Fatalf("WithRateLimiter returned error: %v", err)
	}
	if err := WithRateLimit(0)(&cfg); err != nil {
		t.Fatalf("WithRateLimit returned error: %v", err)
	}
	if cfg.rateLimiter != (transport.RateLimiterConfig{}) {
		t.Fatalf("expected limiter to be disabled, got %#v", cfg.rateLimiter)
	}
}

func TestWithRetryBackoffOverridesDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithRetryBackoff(250*time.Millisecond, 3*time.Second)(&cfg); err != nil {
		t.Fatalf("WithRetryBackoff returned error: %v", err)
	}

	want := request.RetryBackoffConfig{
		BaseDelay:         250 * time.Millisecond,
		MaxDelay:          3 * time.Second,
		RespectRetryAfter: true,
	}
	if cfg.retryBackoff != want {
		t.Fatalf("retryBackoff = %#v, want %#v", cfg.retryBackoff, want)
	}
}

func TestWithRetryBackoffRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithRetryBackoff(0, time.Second)(&cfg); err == nil {
		t.Fatal("expected error for zero base delay")
	}
	if err := WithRetryBackoff(time.Second, 0)(&cfg); err == nil {
		t.Fatal("expected error for zero max delay")
	}
	if err := WithRetryBackoff(2*time.Second, time.Second)(&cfg); err == nil {
		t.Fatal("expected error when max delay is smaller than base delay")
	}
}

func TestWithRetryRespectRetryAfterOverridesDefault(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithRetryRespectRetryAfter(false)(&cfg); err != nil {
		t.Fatalf("WithRetryRespectRetryAfter returned error: %v", err)
	}
	if cfg.retryBackoff.RespectRetryAfter {
		t.Fatal("expected Retry-After handling to be disabled")
	}
}

func TestWithSafeDefaultsAppliesConservativeRetryAndLimiter(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithSafeDefaults()(&cfg); err != nil {
		t.Fatalf("WithSafeDefaults returned error: %v", err)
	}

	if cfg.retry != safeDefaultRetry {
		t.Fatalf("unexpected retry: got %d want %d", cfg.retry, safeDefaultRetry)
	}
	if cfg.rateLimiter.Limit != rate.Limit(safeDefaultRPS) {
		t.Fatalf("unexpected rate limit: got %v want %v", cfg.rateLimiter.Limit, rate.Limit(safeDefaultRPS))
	}
	if cfg.rateLimiter.Burst != safeDefaultRPS {
		t.Fatalf("unexpected burst: got %d want %d", cfg.rateLimiter.Burst, safeDefaultRPS)
	}
}

func TestDefaultAPIKeyHealthConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultAPIKeyHealthConfig()
	if cfg.FailureThreshold != 2 {
		t.Fatalf("unexpected failure threshold: %d", cfg.FailureThreshold)
	}
	if cfg.Cooldown != 30*time.Second {
		t.Fatalf("unexpected cooldown: %s", cfg.Cooldown)
	}
}

func TestWithHealthCheckedAPIKeysSetsProvider(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithHealthCheckedAPIKeys(APIKeyHealthConfig{FailureThreshold: 1, Cooldown: time.Second}, "key-a", "key-b")(&cfg); err != nil {
		t.Fatalf("WithHealthCheckedAPIKeys returned error: %v", err)
	}
	if cfg.apiKeyProvider == nil {
		t.Fatal("expected api key provider")
	}
}

func TestWithHealthCheckedAPIKeysRejectsNegativeConfig(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithHealthCheckedAPIKeys(APIKeyHealthConfig{FailureThreshold: -1}, "key-a")(&cfg); err == nil {
		t.Fatal("expected error for negative failure threshold")
	}
	if err := WithHealthCheckedAPIKeys(APIKeyHealthConfig{Cooldown: -time.Second}, "key-a")(&cfg); err == nil {
		t.Fatal("expected error for negative cooldown")
	}
}

func TestWithCookieJarSetsConfig(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	jar := stubCookieJar{}
	if err := WithCookieJar(jar)(&cfg); err != nil {
		t.Fatalf("WithCookieJar returned error: %v", err)
	}
	if !cfg.cookieJarConfigured {
		t.Fatal("expected cookie jar to be marked as configured")
	}
	if cfg.cookieJar != jar {
		t.Fatalf("cookieJar = %#v, want %#v", cfg.cookieJar, jar)
	}
}

func TestWithCookieJarNilDisablesJar(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	cfg.cookieJar = stubCookieJar{}
	if err := WithCookieJar(nil)(&cfg); err != nil {
		t.Fatalf("WithCookieJar returned error: %v", err)
	}
	if !cfg.cookieJarConfigured {
		t.Fatal("expected cookie jar to be marked as configured")
	}
	if cfg.cookieJar != nil {
		t.Fatalf("expected cookie jar to be nil, got %#v", cfg.cookieJar)
	}
}

func TestWithDefaultCookieJarCreatesJar(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithDefaultCookieJar()(&cfg); err != nil {
		t.Fatalf("WithDefaultCookieJar returned error: %v", err)
	}
	if !cfg.cookieJarConfigured {
		t.Fatal("expected cookie jar to be marked as configured")
	}
	if cfg.cookieJar == nil {
		t.Fatal("expected cookie jar to be created")
	}
}

func TestLastCookieJarOptionWins(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	customJar := stubCookieJar{}
	if err := WithCookieJar(customJar)(&cfg); err != nil {
		t.Fatalf("WithCookieJar returned error: %v", err)
	}
	if err := WithDefaultCookieJar()(&cfg); err != nil {
		t.Fatalf("WithDefaultCookieJar returned error: %v", err)
	}
	if cfg.cookieJar == nil {
		t.Fatal("expected default cookie jar to override custom jar")
	}
	if cfg.cookieJar == customJar {
		t.Fatal("expected default cookie jar to replace custom jar")
	}

	if err := WithCookieJar(nil)(&cfg); err != nil {
		t.Fatalf("WithCookieJar returned error: %v", err)
	}
	if cfg.cookieJar != nil {
		t.Fatalf("expected nil cookie jar to override prior config, got %#v", cfg.cookieJar)
	}
}

func TestWithTrafficPolicyStoresPerClassPolicy(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	jar := stubCookieJar{}
	backoff := DefaultRetryBackoffConfig()
	profile := DefaultPublicStoreHeaderProfileZH()
	selector, err := NewStaticRefererSelector("https://store.steampowered.com/search/")
	if err != nil {
		t.Fatalf("NewStaticRefererSelector returned error: %v", err)
	}
	hook := TransportHookFunc(func(class TrafficClass, base *http.Client) (*http.Client, error) {
		return cloneHTTPClient(base), nil
	})
	if err := WithTrafficPolicy(TrafficClassPublicStorePage, TrafficPolicy{
		CookieJar: jar,
		RateLimiter: &TrafficRateLimiterPolicy{
			Limit: rate.Limit(4),
			Burst: 6,
		},
		Retry: &TrafficRetryPolicy{
			Retry:   2,
			Backoff: backoff,
		},
		HostControl: &TrafficHostControlPolicy{
			RateLimiter: &TrafficRateLimiterPolicy{
				Limit: rate.Limit(2),
				Burst: 3,
			},
			MaxConcurrent: 4,
		},
		SessionControl: &TrafficSessionControlPolicy{
			RateLimiter: &TrafficRateLimiterPolicy{
				Limit: rate.Limit(1),
				Burst: 2,
			},
			MaxConcurrent: 5,
		},
		Cache:           &TrafficCachePolicy{TTL: time.Minute},
		BlockPolicy:     &TrafficBlockPolicy{HTMLSniffBytes: 4096},
		HeaderProfile:   &profile,
		RefererSelector: selector,
		TransportHook:   hook,
	})(&cfg); err != nil {
		t.Fatalf("WithTrafficPolicy returned error: %v", err)
	}

	policy, ok := cfg.trafficPolicies[TrafficClassPublicStorePage]
	if !ok {
		t.Fatal("expected per-class policy to be stored")
	}
	if !policy.cookieJarProvided || policy.cookieJar != jar {
		t.Fatalf("unexpected cookie jar policy: %#v", policy)
	}
	if policy.rateLimiter == nil || policy.rateLimiter.Limit != rate.Limit(4) || policy.rateLimiter.Burst != 6 {
		t.Fatalf("unexpected rate limiter policy: %#v", policy.rateLimiter)
	}
	if policy.retry == nil || policy.retry.Retry != 2 || policy.retry.Backoff != backoff {
		t.Fatalf("unexpected retry policy: %#v", policy.retry)
	}
	if policy.hostControl == nil || policy.hostControl.MaxConcurrent != 4 {
		t.Fatalf("unexpected host control policy: %#v", policy.hostControl)
	}
	if policy.hostControl.RateLimiter == nil || policy.hostControl.RateLimiter.Limit != rate.Limit(2) || policy.hostControl.RateLimiter.Burst != 3 {
		t.Fatalf("unexpected host control limiter: %#v", policy.hostControl)
	}
	if policy.sessionControl == nil || policy.sessionControl.MaxConcurrent != 5 {
		t.Fatalf("unexpected session control policy: %#v", policy.sessionControl)
	}
	if policy.sessionControl.RateLimiter == nil || policy.sessionControl.RateLimiter.Limit != rate.Limit(1) || policy.sessionControl.RateLimiter.Burst != 2 {
		t.Fatalf("unexpected session control limiter: %#v", policy.sessionControl)
	}
	if policy.cache == nil || policy.cache.TTL != time.Minute {
		t.Fatalf("unexpected cache policy: %#v", policy.cache)
	}
	if policy.blockPolicy == nil || policy.blockPolicy.HTMLSniffBytes != 4096 {
		t.Fatalf("unexpected block policy: %#v", policy.blockPolicy)
	}
	if policy.headerProfile == nil || policy.headerProfile.AcceptLanguage != profile.AcceptLanguage {
		t.Fatalf("unexpected header profile: %#v", policy.headerProfile)
	}
	if policy.refererSelector == nil {
		t.Fatal("expected referer selector to be stored")
	}
	if policy.transportHook == nil {
		t.Fatal("expected transport hook to be stored")
	}
}

func TestBuildHTTPClientClonesDefaultTransport(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	first, err := buildHTTPClient(cfg, nil, false)
	if err != nil {
		t.Fatalf("buildHTTPClient returned error: %v", err)
	}
	second, err := buildHTTPClient(cfg, nil, false)
	if err != nil {
		t.Fatalf("buildHTTPClient returned error: %v", err)
	}

	if first.Transport == http.DefaultTransport || second.Transport == http.DefaultTransport {
		t.Fatal("expected default clients to clone http.DefaultTransport")
	}
	if first.Transport == second.Transport {
		t.Fatal("expected each default client to receive its own transport clone")
	}
}

func TestWithTrafficPolicyRejectsUnsupportedClass(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithTrafficPolicy(TrafficClass("unknown"), TrafficPolicy{})(&cfg); err == nil {
		t.Fatal("expected error for unsupported traffic class")
	}
}

func TestWithTrafficPolicyRejectsNegativeHostMaxConcurrent(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	err := WithTrafficPolicy(TrafficClassPublicStorePage, TrafficPolicy{
		HostControl: &TrafficHostControlPolicy{MaxConcurrent: -1},
	})(&cfg)
	if err == nil {
		t.Fatal("expected error for negative host max concurrent")
	}
}

func TestWithTrafficPolicyRejectsNegativeSessionMaxConcurrent(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	err := WithTrafficPolicy(TrafficClassPublicStorePage, TrafficPolicy{
		SessionControl: &TrafficSessionControlPolicy{MaxConcurrent: -1},
	})(&cfg)
	if err == nil {
		t.Fatal("expected error for negative session max concurrent")
	}
}

func TestWithTrafficPolicyRejectsInvalidHostRateLimiter(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	err := WithTrafficPolicy(TrafficClassPublicStorePage, TrafficPolicy{
		HostControl: &TrafficHostControlPolicy{
			RateLimiter: &TrafficRateLimiterPolicy{Limit: rate.Limit(1), Burst: 0},
		},
	})(&cfg)
	if err == nil {
		t.Fatal("expected error for invalid host rate limiter")
	}
}

func TestWithTrafficPolicyRejectsInvalidSessionRateLimiter(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	err := WithTrafficPolicy(TrafficClassPublicStorePage, TrafficPolicy{
		SessionControl: &TrafficSessionControlPolicy{
			RateLimiter: &TrafficRateLimiterPolicy{Limit: rate.Limit(1), Burst: 0},
		},
	})(&cfg)
	if err == nil {
		t.Fatal("expected error for invalid session rate limiter")
	}
}

func TestWithTrafficPolicyRejectsInvalidCacheTTL(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	err := WithTrafficPolicy(TrafficClassPublicStorePage, TrafficPolicy{
		Cache: &TrafficCachePolicy{TTL: 0},
	})(&cfg)
	if err == nil {
		t.Fatal("expected error for zero cache ttl")
	}
}

func TestWithTrafficPolicyRejectsBlockPolicyForOfficialAPI(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	err := WithTrafficPolicy(TrafficClassOfficialAPI, TrafficPolicy{
		BlockPolicy: &TrafficBlockPolicy{},
	})(&cfg)
	if err == nil {
		t.Fatal("expected error for official api block policy")
	}
}

func TestWithTrafficPolicyRejectsNegativeBlockHTMLSniffBytes(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	err := WithTrafficPolicy(TrafficClassPublicStorePage, TrafficPolicy{
		BlockPolicy: &TrafficBlockPolicy{HTMLSniffBytes: -1},
	})(&cfg)
	if err == nil {
		t.Fatal("expected error for negative html sniff bytes")
	}
}

func TestWithRequestSessionKeyStoresTrimmedKey(t *testing.T) {
	t.Parallel()

	ctx := WithRequestSessionKey(nil, "  session-a  ")
	key, ok := traffic.RequestSessionKeyFromContext(ctx)
	if !ok {
		t.Fatal("expected request session key")
	}
	if key != "session-a" {
		t.Fatalf("unexpected request session key: %q", key)
	}
}

func TestWithRequestSessionKeyTreatsBlankAsUnset(t *testing.T) {
	t.Parallel()

	base := context.Background()
	ctx := WithRequestSessionKey(base, "   ")
	if ctx != base {
		t.Fatal("expected blank request session key to keep original context")
	}
	if _, ok := traffic.RequestSessionKeyFromContext(ctx); ok {
		t.Fatal("did not expect request session key")
	}
}

func TestWithRequestSessionKeyDoesNotAffectProxySessionKey(t *testing.T) {
	t.Parallel()

	ctx := WithProxySessionKey(context.Background(), "proxy-a")
	ctx = WithRequestSessionKey(ctx, "request-a")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}

	if got := proxySessionKeyFromRequest(req); got != "proxy-a" {
		t.Fatalf("unexpected proxy session key: %q", got)
	}
	if key, ok := traffic.RequestSessionKeyFromContext(req.Context()); !ok || key != "request-a" {
		t.Fatalf("unexpected request session key: %q / %v", key, ok)
	}
}
