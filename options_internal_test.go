package steam

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/GoFurry/steam-go/internal/request"
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
