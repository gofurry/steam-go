package steam_test

import (
	"testing"

	steam "github.com/GoFurry/steam-go"
)

func TestRedactSensitiveURLRemovesCredentials(t *testing.T) {
	t.Parallel()

	got := steam.RedactSensitiveURL("https://user:secret@example.com/path?foo=bar&key=api-key&access_token=token")
	want := "https://example.com/path?foo=bar"
	if got != want {
		t.Fatalf("unexpected redacted url: got %q want %q", got, want)
	}
}

func TestRedactSensitiveURLLeavesInvalidURLUntouched(t *testing.T) {
	t.Parallel()

	raw := "://not-a-url"
	if got := steam.RedactSensitiveURL(raw); got != raw {
		t.Fatalf("unexpected fallback for invalid url: got %q want %q", got, raw)
	}
}
