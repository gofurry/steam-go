package markup

import (
	"strings"
	"testing"
)

func TestSteamBBCodeToHTMLConvertsCommonTags(t *testing.T) {
	t.Parallel()

	got, err := SteamBBCodeToHTML(`[b]bold[/b] [i]italic[/i] [url=https://example.test]link[/url]`)
	if err != nil {
		t.Fatalf("SteamBBCodeToHTML returned error: %v", err)
	}
	for _, want := range []string{
		"<strong>bold</strong>",
		"<em>italic</em>",
		`href="https://example.test"`,
		">link</a>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in %q", want, got)
		}
	}
}

func TestSteamBBCodeToHTMLConvertsSteamImagesAndLists(t *testing.T) {
	t.Parallel()

	got, err := SteamBBCodeToHTML(`[img]{STEAM_CLAN_IMAGE}/abc.png[/img][list][*]one[/*][*]two[/*][/list]`)
	if err != nil {
		t.Fatalf("SteamBBCodeToHTML returned error: %v", err)
	}
	for _, want := range []string{
		`src="https://clan.fastly.steamstatic.com/images/abc.png"`,
		"<ul>",
		"<li>one</li>",
		"<li>two</li>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in %q", want, got)
		}
	}
}

func TestSteamBBCodeToHTMLHandlesSteamEscapedTextAndURLs(t *testing.T) {
	t.Parallel()

	got, err := SteamBBCodeToHTML(`[p]\[ PATCH ][/p][url=\"https://example.test/news\"]notes[/url]`)
	if err != nil {
		t.Fatalf("SteamBBCodeToHTML returned error: %v", err)
	}
	for _, want := range []string{
		"[ PATCH ]",
		`href="https://example.test/news"`,
		">notes</a>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in %q", want, got)
		}
	}
}

func TestCleanSteamContentSanitizesDangerousHTML(t *testing.T) {
	t.Parallel()

	got, err := CleanSteamContent(`<script>alert(1)</script><img src="https://example.test/a.png" onerror="alert(1)"><a href="javascript:alert(1)">bad</a>`)
	if err != nil {
		t.Fatalf("CleanSteamContent returned error: %v", err)
	}
	for _, forbidden := range []string{"<script", "onerror", "javascript:"} {
		if strings.Contains(strings.ToLower(got), forbidden) {
			t.Fatalf("expected %q to be removed from %q", forbidden, got)
		}
	}
	if !strings.Contains(got, `<img src="https://example.test/a.png"`) {
		t.Fatalf("expected safe image to remain, got %q", got)
	}
}

func TestPlainTextAndSummary(t *testing.T) {
	t.Parallel()

	text, err := PlainText(`[h1]Title[/h1] [b]Hello[/b]<br>world`)
	if err != nil {
		t.Fatalf("PlainText returned error: %v", err)
	}
	if text != "Title Hello world" {
		t.Fatalf("unexpected plain text: %q", text)
	}

	summary, err := Summary(`[b]abcdef[/b]`, 3)
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}
	if summary != "abc" {
		t.Fatalf("unexpected summary: %q", summary)
	}
}
