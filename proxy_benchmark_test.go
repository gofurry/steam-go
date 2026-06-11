package steam_test

import (
	"net/http"
	"testing"

	steam "github.com/gofurry/steam-go"
)

func BenchmarkStickyProxySelector(b *testing.B) {
	base, err := steam.NewRoundRobinProxySelector(
		"http://127.0.0.1:7897",
		"http://127.0.0.1:7898",
	)
	if err != nil {
		b.Fatalf("NewRoundRobinProxySelector returned error: %v", err)
	}
	selector := steam.NewStickyProxySelector(base)
	req, err := http.NewRequest(http.MethodGet, "https://store.steampowered.com/app/10", nil)
	if err != nil {
		b.Fatalf("http.NewRequest returned error: %v", err)
	}
	req = req.WithContext(steam.WithProxySessionKey(req.Context(), "session-a"))

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := selector.Next(req); err != nil {
			b.Fatalf("Next returned error: %v", err)
		}
	}
}
