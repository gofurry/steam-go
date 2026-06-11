package transport

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"testing"

	"github.com/gofurry/steam-go/internal/traffic"
)

func BenchmarkTransportContextCookieJar(b *testing.B) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		b.Fatalf("cookiejar.New returned error: %v", err)
	}
	client := New(&http.Client{
		Jar: jar,
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return okResponse(), nil
		}),
	}, ClientConfig{})
	ctx := traffic.WithCookieJar(context.Background(), jar)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.steampowered.com/one", nil)
	if err != nil {
		b.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resp, err := client.Do(ctx, req)
		if err != nil {
			b.Fatalf("Do returned error: %v", err)
		}
		_ = resp.Body.Close()
	}
}

func BenchmarkRequestControlHighCardinality(b *testing.B) {
	client := New(&http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return okResponse(), nil
		}),
	}, ClientConfig{
		HostControl: RequestControlConfig{MaxConcurrent: 1},
	})

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rawURL := "https://host-" + strconv.Itoa(i) + ".example/one"
		if err := doTransportRequest(client, context.Background(), rawURL); err != nil {
			b.Fatalf("request returned error: %v", err)
		}
	}
}
