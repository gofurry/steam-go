package steam_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	steam "github.com/gofurry/steam-go"
)

func TestRedactionGolden(t *testing.T) {
	t.Parallel()

	headers := steam.RedactSensitiveHeaders(http.Header{
		"Authorization": []string{"Bearer fixture-token"},
		"Cookie":        []string{"steamLoginSecure=fixture"},
		"Content-Type":  []string{"application/json"},
	})
	got := strings.Join([]string{
		"url=" + steam.RedactSensitiveURL("https://store.steampowered.com/api/appdetails?appids=550&key=fixture-key&cc=US"),
		"proxy=" + steam.RedactSensitiveURL("http://user:password@proxy.example:8080"),
		"authorization=" + headers.Get("Authorization"),
		"cookie=" + headers.Get("Cookie"),
		"content-type=" + headers.Get("Content-Type"),
	}, "\n") + "\n"

	want, err := os.ReadFile(filepath.Join("testdata", "golden", "redaction.txt"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != normalizeGoldenLineEndings(string(want)) {
		t.Fatalf("redaction golden mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func normalizeGoldenLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
