package steam_test

import (
	"net/http"
	"strings"
	"testing"

	steam "github.com/gofurry/steam-go"
)

func TestRedactSensitiveURLRemovesCredentials(t *testing.T) {
	t.Parallel()

	got := steam.RedactSensitiveURL("https://user:secret@example.com/path?foo=bar&key=api-key&access_token=token&refresh_token=refresh&steamLoginSecure=secure&sessionid=session&webapi_token=web&loyalty_webapi_token=loyalty")
	want := "https://example.com/path?foo=bar"
	if got != want {
		t.Fatalf("unexpected redacted url: got %q want %q", got, want)
	}
}

func TestRedactSensitiveURLRemovesCredentialLikeQueryCaseInsensitively(t *testing.T) {
	t.Parallel()

	got := steam.RedactSensitiveURL("https://example.com/path?Access_Token=token&SteamLoginSecure=secure&ok=1")
	want := "https://example.com/path?ok=1"
	if got != want {
		t.Fatalf("unexpected redacted url: got %q want %q", got, want)
	}
}

func TestRedactSensitiveURLRemovesProxyUserinfo(t *testing.T) {
	t.Parallel()

	got := steam.RedactSensitiveURL("http://user:pass@proxy.example:8080")
	want := "http://proxy.example:8080"
	if got != want {
		t.Fatalf("unexpected redacted proxy url: got %q want %q", got, want)
	}
}

func TestRedactSensitiveURLRedactsRedirectFinalURL(t *testing.T) {
	t.Parallel()

	got := steam.RedactSensitiveURL("https://store.steampowered.com/login/?return_to=%2F&sessionid=session&refresh_token=refresh")
	want := "https://store.steampowered.com/login/?return_to=%2F"
	if got != want {
		t.Fatalf("unexpected redacted redirect url: got %q want %q", got, want)
	}
}

func TestRedactSensitiveURLLeavesInvalidURLUntouched(t *testing.T) {
	t.Parallel()

	raw := "://not-a-url"
	if got := steam.RedactSensitiveURL(raw); got != raw {
		t.Fatalf("unexpected fallback for invalid url: got %q want %q", got, raw)
	}
}

func TestRedactSensitiveURLFallbackRedactsMalformedURLLikeInput(t *testing.T) {
	t.Parallel()

	raw := "http://example.com/%zz?access_token=abc123&x=1 steamLoginSecure=hidden-cookie;refresh_token=hidden-refresh"
	got := steam.RedactSensitiveURL(raw)
	if strings.Contains(got, "abc123") || strings.Contains(got, "hidden-cookie") || strings.Contains(got, "hidden-refresh") {
		t.Fatalf("fallback redaction leaked sensitive values: %q", got)
	}
	if !strings.Contains(got, "access_token=[REDACTED]") ||
		!strings.Contains(got, "steamLoginSecure=[REDACTED]") ||
		!strings.Contains(got, "refresh_token=[REDACTED]") ||
		!strings.Contains(got, "x=1") {
		t.Fatalf("unexpected fallback redaction: %q", got)
	}
}

func TestRedactSensitiveHeaderValueRedactsCredentialBearingHeaders(t *testing.T) {
	t.Parallel()

	tests := []string{
		"Authorization",
		"Proxy-Authorization",
		"Cookie",
		"Set-Cookie",
		"X-WebAPI-Key",
		"X-API-Key",
	}
	for _, name := range tests {
		if got := steam.RedactSensitiveHeaderValue(name, "secret"); got != "[REDACTED]" {
			t.Fatalf("RedactSensitiveHeaderValue(%q) = %q, want [REDACTED]", name, got)
		}
	}
}

func TestRedactSensitiveHeaderValueLeavesSafeHeaders(t *testing.T) {
	t.Parallel()

	if got := steam.RedactSensitiveHeaderValue("Content-Type", "application/json"); got != "application/json" {
		t.Fatalf("unexpected safe header value: %q", got)
	}
}

func TestRedactSensitiveHeadersClonesAndRedactsValues(t *testing.T) {
	t.Parallel()

	header := http.Header{
		"Authorization": []string{"Bearer secret"},
		"Cookie":        []string{"steamLoginSecure=secret"},
		"Content-Type":  []string{"application/json"},
	}

	got := steam.RedactSensitiveHeaders(header)
	if got.Get("Authorization") != "[REDACTED]" {
		t.Fatalf("Authorization = %q, want [REDACTED]", got.Get("Authorization"))
	}
	if got.Get("Cookie") != "[REDACTED]" {
		t.Fatalf("Cookie = %q, want [REDACTED]", got.Get("Cookie"))
	}
	if got.Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got.Get("Content-Type"))
	}

	got.Set("Content-Type", "text/plain")
	if header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected original header to remain unchanged, got %q", header.Get("Content-Type"))
	}
	if header.Get("Authorization") != "Bearer secret" {
		t.Fatalf("expected original Authorization to remain unchanged, got %q", header.Get("Authorization"))
	}
}

func TestRedactSensitiveHeadersNil(t *testing.T) {
	t.Parallel()

	if got := steam.RedactSensitiveHeaders(nil); got != nil {
		t.Fatalf("expected nil header, got %#v", got)
	}
}
