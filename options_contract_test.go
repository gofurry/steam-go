package steam

import (
	"testing"

	"github.com/GoFurry/steam-go/internal/transport"
	"golang.org/x/time/rate"
)

func TestWithSafeDefaultsCanBeOverriddenByExplicitOptions(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithSafeDefaults()(&cfg); err != nil {
		t.Fatalf("WithSafeDefaults returned error: %v", err)
	}
	if err := WithRetry(4)(&cfg); err != nil {
		t.Fatalf("WithRetry returned error: %v", err)
	}
	if err := WithRateLimiter(rate.Limit(9), 11)(&cfg); err != nil {
		t.Fatalf("WithRateLimiter returned error: %v", err)
	}

	if cfg.retry != 4 {
		t.Fatalf("unexpected retry: got %d want %d", cfg.retry, 4)
	}
	wantLimiter := transport.RateLimiterConfig{
		Limit: rate.Limit(9),
		Burst: 11,
	}
	if cfg.rateLimiter != wantLimiter {
		t.Fatalf("unexpected rate limiter: got %#v want %#v", cfg.rateLimiter, wantLimiter)
	}
}

func TestWithSafeDefaultsAppliedLastResetsEarlierRetryAndLimiter(t *testing.T) {
	t.Parallel()

	cfg := defaultClientConfig()
	if err := WithRetry(4)(&cfg); err != nil {
		t.Fatalf("WithRetry returned error: %v", err)
	}
	if err := WithRateLimiter(rate.Limit(9), 11)(&cfg); err != nil {
		t.Fatalf("WithRateLimiter returned error: %v", err)
	}
	if err := WithSafeDefaults()(&cfg); err != nil {
		t.Fatalf("WithSafeDefaults returned error: %v", err)
	}

	if cfg.retry != safeDefaultRetry {
		t.Fatalf("unexpected retry: got %d want %d", cfg.retry, safeDefaultRetry)
	}
	wantLimiter := transport.RateLimiterConfig{
		Limit: rate.Limit(safeDefaultRPS),
		Burst: safeDefaultRPS,
	}
	if cfg.rateLimiter != wantLimiter {
		t.Fatalf("unexpected rate limiter: got %#v want %#v", cfg.rateLimiter, wantLimiter)
	}
}
