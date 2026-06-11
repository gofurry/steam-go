package websession

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func TestWebCookieSnapshotRoundTrip(t *testing.T) {
	t.Parallel()

	storeURL := mustParseURL(t, "https://store.steampowered.com/")
	communityURL := mustParseURL(t, "https://steamcommunity.com/")
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}
	jar.SetCookies(storeURL, []*http.Cookie{
		{Name: "sessionid", Value: "store-session", Path: "/"},
		{Name: "steamLoginSecure", Value: "store-secure", Path: "/"},
	})
	jar.SetCookies(communityURL, []*http.Cookie{
		{Name: "steamLoginSecure", Value: "community-secure", Path: "/"},
	})

	var buf bytes.Buffer
	err = SaveWebCookieResultJSON(&buf, &WebCookieResult{
		Jar:       jar,
		SessionID: "store-session",
		SteamID:   "76561198000000001",
		Domains:   []string{"steamcommunity.com", "store.steampowered.com"},
	})
	if err != nil {
		t.Fatalf("SaveWebCookieResultJSON returned error: %v", err)
	}
	if bytes.Contains(buf.Bytes(), []byte("password")) {
		t.Fatal("snapshot should not contain password-like fields")
	}

	restored, err := LoadWebCookieResultJSON(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("LoadWebCookieResultJSON returned error: %v", err)
	}
	if restored.SteamID != "76561198000000001" || restored.SessionID != "store-session" {
		t.Fatalf("unexpected restored metadata: %#v", restored)
	}
	assertCookieValue(t, restored.Jar, storeURL, "sessionid", "store-session")
	assertCookieValue(t, restored.Jar, storeURL, "steamLoginSecure", "store-secure")
	assertCookieValue(t, restored.Jar, communityURL, "steamLoginSecure", "community-secure")
}

func TestImportWebCookieSnapshotValidation(t *testing.T) {
	t.Parallel()

	_, err := ImportWebCookieSnapshot(WebCookieSnapshot{Version: 99})
	if err == nil {
		t.Fatal("expected unsupported version error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) || clientErr.Code != ErrorCodeDecode {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = ImportWebCookieSnapshot(WebCookieSnapshot{Version: webCookieSnapshotVersion})
	if err == nil {
		t.Fatal("expected empty snapshot error")
	}
}

func TestExportWebCookieSnapshotValidation(t *testing.T) {
	t.Parallel()

	if _, err := ExportWebCookieSnapshot(nil); err == nil {
		t.Fatal("expected nil result error")
	}
	if _, err := ExportWebCookieSnapshot(&WebCookieResult{}); err == nil {
		t.Fatal("expected nil jar error")
	}
}

func assertCookieValue(t *testing.T, jar http.CookieJar, rawURL *url.URL, name, want string) {
	t.Helper()

	for _, cookie := range jar.Cookies(rawURL) {
		if cookie.Name == name {
			if cookie.Value != want {
				t.Fatalf("unexpected cookie %s=%q want %q", name, cookie.Value, want)
			}
			return
		}
	}
	t.Fatalf("expected cookie %s", name)
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}
	return parsed
}
