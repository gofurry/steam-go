package steam_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	steam "github.com/gofurry/steam-go"
)

func BenchmarkRequestObserver(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	b.Run("no_observer", func(b *testing.B) {
		benchmarkRawRequest(b, server.URL)
	})

	b.Run("noop_observer", func(b *testing.B) {
		benchmarkRawRequest(
			b,
			server.URL,
			steam.WithRequestObserver(steam.RequestObserverFunc(func(steam.RequestEvent) {})),
		)
	})

	b.Run("counter_observer", func(b *testing.B) {
		var events atomic.Int64
		benchmarkRawRequest(
			b,
			server.URL,
			steam.WithRequestObserver(steam.RequestObserverFunc(func(steam.RequestEvent) {
				events.Add(1)
			})),
		)
		if got := events.Load(); got != int64(b.N) {
			b.Fatalf("observer saw %d events, want %d", got, b.N)
		}
	})
}

func benchmarkRawRequest(b *testing.B, baseURL string, opts ...steam.Option) {
	b.Helper()

	client, err := steam.NewClient(opts...)
	if err != nil {
		b.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/ok", nil)
		if err != nil {
			b.Fatalf("NewRequestWithContext returned error: %v", err)
		}
		if _, err := client.DoRawHTTPRequest(context.Background(), req, &steam.RawHTTPRequestOptions{TrafficClass: steam.TrafficClassOfficialAPI}); err != nil {
			b.Fatalf("DoRawHTTPRequest returned error: %v", err)
		}
	}
}
