package auth

import (
	"net/http"
	"testing"
	"time"
)

func TestHealthCheckedRoundRobinKeyProviderSkipsCoolingKeys(t *testing.T) {
	t.Parallel()

	provider, err := NewHealthCheckedRoundRobinKeyProvider(
		KeyHealthConfig{FailureThreshold: 1, Cooldown: 100 * time.Millisecond},
		"key-a",
		"key-b",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinKeyProvider returned error: %v", err)
	}

	keyProvider := provider.(*healthCheckedRoundRobinKeyProvider)
	first, err := keyProvider.Next(nil)
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if first != "key-a" {
		t.Fatalf("unexpected first key: %s", first)
	}

	keyProvider.ReportAPIKeyResult(nil, first, http.StatusUnauthorized, nil)

	second, err := keyProvider.Next(nil)
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if second != "key-b" {
		t.Fatalf("expected cooled key to be skipped, got %s", second)
	}
}

func TestHealthCheckedRoundRobinKeyProviderResetsAfterSuccess(t *testing.T) {
	t.Parallel()

	provider, err := NewHealthCheckedRoundRobinKeyProvider(
		KeyHealthConfig{FailureThreshold: 2, Cooldown: time.Second},
		"key-a",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinKeyProvider returned error: %v", err)
	}

	keyProvider := provider.(*healthCheckedRoundRobinKeyProvider)
	keyProvider.ReportAPIKeyResult(nil, "key-a", http.StatusUnauthorized, nil)
	keyProvider.ReportAPIKeyResult(nil, "key-a", http.StatusOK, nil)
	keyProvider.ReportAPIKeyResult(nil, "key-a", http.StatusUnauthorized, nil)

	if state := keyProvider.states[0]; !state.cooldownUntil.IsZero() || state.failureScore != 1 {
		t.Fatalf("unexpected state after reset path: %+v", state)
	}
}
