package steamid_test

import (
	"strings"
	"testing"

	"github.com/gofurry/steam-go/internal/steamid"
)

func TestValidateSteamID64(t *testing.T) {
	t.Parallel()

	got, err := steamid.ValidateSteamID64(" 76561198000000000 ")
	if err != nil {
		t.Fatalf("ValidateSteamID64 returned error: %v", err)
	}
	if got != "76561198000000000" {
		t.Fatalf("unexpected normalized steam id: %q", got)
	}
}

func TestValidateSteamID64RejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"   ",
		"76561198000000000x",
		"18446744073709551616",
	}
	for _, value := range tests {
		if _, err := steamid.ValidateSteamID64(value); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}

func TestValidateNumericIDUsesCustomName(t *testing.T) {
	t.Parallel()

	_, err := steamid.ValidateNumericID("context id", "abc")
	if err == nil || !strings.Contains(err.Error(), "context id") {
		t.Fatalf("expected named validation error, got %v", err)
	}
}
