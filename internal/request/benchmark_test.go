package request

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func BenchmarkCacheHitMiss(b *testing.B) {
	runtime := NewMemoryCacheRuntime(time.Minute, nil)
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	req := newCacheTestRequestForBenchmark(b, "https://store.steampowered.com/app/10")
	runtime.store(req, resp, newCacheTestResult("body"), time.Now())

	b.Run("hit", func(b *testing.B) {
		now := time.Now()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = runtime.lookup(req, now)
		}
	})

	b.Run("miss", func(b *testing.B) {
		now := time.Now()
		missReq := newCacheTestRequestForBenchmark(b, "https://store.steampowered.com/app/20")
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = runtime.lookup(missReq, now)
		}
	})
}

func BenchmarkRawHTTPReadLimit(b *testing.B) {
	body := strings.Repeat("x", 64<<10)
	policy := ExecutionPolicy{
		RetryBackoff: DefaultRetryBackoffConfig(),
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		}),
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := newCacheTestRequestForBenchmark(b, "https://cdn.steamstatic.com/file")
		if _, err := ExecuteRawHTTPRequest(context.Background(), req, int64(len(body)+1), policy, nil); err != nil {
			b.Fatalf("ExecuteRawHTTPRequest returned error: %v", err)
		}
	}
}

func newCacheTestRequestForBenchmark(b *testing.B, rawURL string) *http.Request {
	b.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		b.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	return req
}
