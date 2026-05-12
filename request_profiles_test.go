package steam

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestDefaultPublicStoreHeaderProfiles(t *testing.T) {
	t.Parallel()

	zh := DefaultPublicStoreHeaderProfileZH()
	if zh.UserAgent == "" || zh.Accept == "" || zh.AcceptLanguage == "" || zh.AcceptEncoding == "" {
		t.Fatal("expected zh header profile defaults to be populated")
	}
	en := DefaultPublicStoreHeaderProfileEN()
	if en.UserAgent == "" || en.Accept == "" || en.AcceptLanguage == "" || en.AcceptEncoding == "" {
		t.Fatal("expected en header profile defaults to be populated")
	}
	if zh.AcceptLanguage == en.AcceptLanguage {
		t.Fatal("expected zh and en profiles to differ by language")
	}
}

func TestBuildRequestPreparerAppliesHeaderProfileDefaults(t *testing.T) {
	t.Parallel()

	profile := DefaultPublicStoreHeaderProfileZH()
	req, err := http.NewRequest(http.MethodGet, "https://store.steampowered.com/app/570/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest returned error: %v", err)
	}
	req.Header.Set("User-Agent", sdkDefaultUserAgent)

	preparer := buildRequestPreparer(&profile, nil)
	if err := preparer(req); err != nil {
		t.Fatalf("preparer returned error: %v", err)
	}

	if got := req.Header.Get("User-Agent"); got != profile.UserAgent {
		t.Fatalf("unexpected user agent: %q", got)
	}
	if got := req.Header.Get("Accept-Language"); got != profile.AcceptLanguage {
		t.Fatalf("unexpected accept-language: %q", got)
	}
}

func TestBuildRequestPreparerPreservesExplicitHeaders(t *testing.T) {
	t.Parallel()

	profile := DefaultPublicStoreHeaderProfileZH()
	req, err := http.NewRequest(http.MethodGet, "https://store.steampowered.com/app/570/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest returned error: %v", err)
	}
	req.Header.Set("User-Agent", "custom-agent")
	req.Header.Set("Accept-Language", "ja-JP")

	preparer := buildRequestPreparer(&profile, nil)
	if err := preparer(req); err != nil {
		t.Fatalf("preparer returned error: %v", err)
	}

	if got := req.Header.Get("User-Agent"); got != "custom-agent" {
		t.Fatalf("expected explicit user agent to win, got %q", got)
	}
	if got := req.Header.Get("Accept-Language"); got != "ja-JP" {
		t.Fatalf("expected explicit accept-language to win, got %q", got)
	}
}

func TestNewStaticRefererSelectorRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	if _, err := NewStaticRefererSelector("://bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewRoutingRefererSelectorRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	if _, err := NewRoutingRefererSelector(RefererRoute{
		Host:       "store.steampowered.com",
		RefererURL: "://bad",
	}); err == nil {
		t.Fatal("expected error")
	}
}

func TestStaticRefererSelectorAppliesReferer(t *testing.T) {
	t.Parallel()

	selector, err := NewStaticRefererSelector("https://store.steampowered.com/")
	if err != nil {
		t.Fatalf("NewStaticRefererSelector returned error: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://store.steampowered.com/app/570/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest returned error: %v", err)
	}

	preparer := buildRequestPreparer(nil, selector)
	if err := preparer(req); err != nil {
		t.Fatalf("preparer returned error: %v", err)
	}
	if got := req.Header.Get("Referer"); got != "https://store.steampowered.com/" {
		t.Fatalf("unexpected referer: %q", got)
	}
}

func TestRoutingRefererSelectorAppliesByHostAndPath(t *testing.T) {
	t.Parallel()

	selector, err := NewRoutingRefererSelector(
		RefererRoute{
			Host:       "store.steampowered.com",
			PathPrefix: "/app/",
			RefererURL: "https://store.steampowered.com/search/",
		},
	)
	if err != nil {
		t.Fatalf("NewRoutingRefererSelector returned error: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://store.steampowered.com/app/570/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest returned error: %v", err)
	}

	preparer := buildRequestPreparer(nil, selector)
	if err := preparer(req); err != nil {
		t.Fatalf("preparer returned error: %v", err)
	}
	if got := req.Header.Get("Referer"); got != "https://store.steampowered.com/search/" {
		t.Fatalf("unexpected referer: %q", got)
	}
}

func TestContextRefererSelectorPrefersContextSource(t *testing.T) {
	t.Parallel()

	fallback, err := NewStaticRefererSelector("https://store.steampowered.com/")
	if err != nil {
		t.Fatalf("NewStaticRefererSelector returned error: %v", err)
	}
	selector := NewContextRefererSelector(fallback)
	req, err := http.NewRequestWithContext(
		WithRefererSource(context.Background(), "https://store.steampowered.com/search/?term=bg3"),
		http.MethodGet,
		"https://store.steampowered.com/app/1086940/",
		nil,
	)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}

	preparer := buildRequestPreparer(nil, selector)
	if err := preparer(req); err != nil {
		t.Fatalf("preparer returned error: %v", err)
	}
	if got := req.Header.Get("Referer"); got != "https://store.steampowered.com/search/?term=bg3" {
		t.Fatalf("unexpected referer: %q", got)
	}
}

func TestContextRefererSelectorFallsBackWhenSourceMissing(t *testing.T) {
	t.Parallel()

	fallback, err := NewStaticRefererSelector("https://store.steampowered.com/")
	if err != nil {
		t.Fatalf("NewStaticRefererSelector returned error: %v", err)
	}
	selector := NewContextRefererSelector(fallback)
	req, err := http.NewRequest(http.MethodGet, "https://store.steampowered.com/app/570/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest returned error: %v", err)
	}

	preparer := buildRequestPreparer(nil, selector)
	if err := preparer(req); err != nil {
		t.Fatalf("preparer returned error: %v", err)
	}
	if got := req.Header.Get("Referer"); got != "https://store.steampowered.com/" {
		t.Fatalf("unexpected referer: %q", got)
	}
}

func TestWithRefererSourceTreatsBlankAsUnset(t *testing.T) {
	t.Parallel()

	base := context.Background()
	ctx := WithRefererSource(base, "   ")
	if ctx != base {
		t.Fatal("expected blank referer source to keep original context")
	}
}

func TestContextRefererSelectorReturnsValidationErrorFromContextSource(t *testing.T) {
	t.Parallel()

	selector := NewContextRefererSelector(nil)
	req, err := http.NewRequestWithContext(
		WithRefererSource(context.Background(), "://bad"),
		http.MethodGet,
		"https://store.steampowered.com/app/570/",
		nil,
	)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}

	preparer := buildRequestPreparer(nil, selector)
	err = preparer(req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "referer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRefererSelectorPreservesExplicitHeader(t *testing.T) {
	t.Parallel()

	selector, err := NewStaticRefererSelector("https://store.steampowered.com/")
	if err != nil {
		t.Fatalf("NewStaticRefererSelector returned error: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://store.steampowered.com/app/570/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest returned error: %v", err)
	}
	req.Header.Set("Referer", "https://example.com/custom")

	preparer := buildRequestPreparer(nil, selector)
	if err := preparer(req); err != nil {
		t.Fatalf("preparer returned error: %v", err)
	}
	if got := req.Header.Get("Referer"); got != "https://example.com/custom" {
		t.Fatalf("expected explicit referer to win, got %q", got)
	}
}
