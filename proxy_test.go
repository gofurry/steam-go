package steam_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	steam "github.com/GoFurry/steam-go"
	"github.com/GoFurry/steam-go/internal/traffic"
)

func TestNewStaticProxySelector(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewStaticProxySelector("http://127.0.0.1:7897")
	if err != nil {
		t.Fatalf("NewStaticProxySelector returned error: %v", err)
	}
	if selector == nil {
		t.Fatal("expected selector")
	}

	proxyURL, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if got := proxyURL.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected proxy url: %s", got)
	}
}

func TestNewStaticProxySelectorEmptyReturnsNil(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewStaticProxySelector("  ")
	if err != nil {
		t.Fatalf("NewStaticProxySelector returned error: %v", err)
	}
	if selector != nil {
		t.Fatal("expected nil selector")
	}
}

func TestNewRoundRobinProxySelector(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewRoundRobinProxySelector(
		"http://127.0.0.1:7897",
		"http://127.0.0.1:7898",
	)
	if err != nil {
		t.Fatalf("NewRoundRobinProxySelector returned error: %v", err)
	}
	if selector == nil {
		t.Fatal("expected selector")
	}

	first, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("first Next returned error: %v", err)
	}
	second, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("second Next returned error: %v", err)
	}
	third, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("third Next returned error: %v", err)
	}

	if got := first.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected first proxy: %s", got)
	}
	if got := second.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected second proxy: %s", got)
	}
	if got := third.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected third proxy: %s", got)
	}
}

func TestNewRoundRobinProxySelectorIgnoresEmptyValues(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewRoundRobinProxySelector("", "http://127.0.0.1:7897", " ")
	if err != nil {
		t.Fatalf("NewRoundRobinProxySelector returned error: %v", err)
	}
	if selector == nil {
		t.Fatal("expected selector")
	}

	proxyURL, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if got := proxyURL.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected proxy url: %s", got)
	}
}

func TestNewRoundRobinProxySelectorRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	if _, err := steam.NewRoundRobinProxySelector("://bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultProxyHealthConfig(t *testing.T) {
	t.Parallel()

	cfg := steam.DefaultProxyHealthConfig()
	if cfg.FailureThreshold != 2 {
		t.Fatalf("unexpected failure threshold: %d", cfg.FailureThreshold)
	}
	if cfg.Cooldown != 30*time.Second {
		t.Fatalf("unexpected cooldown: %s", cfg.Cooldown)
	}
}

func TestNewHealthCheckedRoundRobinProxySelector(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{},
		"http://127.0.0.1:7897",
		"http://127.0.0.1:7898",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}
	if selector == nil {
		t.Fatal("expected selector")
	}

	first, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("first Next returned error: %v", err)
	}
	second, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("second Next returned error: %v", err)
	}

	if got := first.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected first proxy: %s", got)
	}
	if got := second.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected second proxy: %s", got)
	}
}

func TestNewHealthCheckedRoundRobinProxySelectorEmptyReturnsNil(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(steam.ProxyHealthConfig{}, "  ")
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}
	if selector != nil {
		t.Fatal("expected nil selector")
	}
}

func TestNewHealthCheckedRoundRobinProxySelectorRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	if _, err := steam.NewHealthCheckedRoundRobinProxySelector(steam.ProxyHealthConfig{}, "://bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewHealthCheckedRoundRobinProxySelectorRejectsNegativeConfig(t *testing.T) {
	t.Parallel()

	if _, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{FailureThreshold: -1},
		"http://127.0.0.1:7897",
	); err == nil {
		t.Fatal("expected error for negative failure threshold")
	}
	if _, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{Cooldown: -time.Second},
		"http://127.0.0.1:7897",
	); err == nil {
		t.Fatal("expected error for negative cooldown")
	}
}

func TestWithProxySessionKeyStoresTrimmedKey(t *testing.T) {
	t.Parallel()

	ctx := steam.WithProxySessionKey(context.Background(), "  session-a  ")
	req := mustRequestWithContext(t, ctx, "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
		},
	}
	selector := steam.NewStickyProxySelector(base)

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if base.calls.Load() != 1 {
		t.Fatalf("expected one base call, got %d", base.calls.Load())
	}
	if got := first.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected first proxy: %s", got)
	}
	if got := second.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected second proxy: %s", got)
	}
}

func TestNewStickyProxySelectorNilBaseReturnsNil(t *testing.T) {
	t.Parallel()

	if selector := steam.NewStickyProxySelector(nil); selector != nil {
		t.Fatal("expected nil selector")
	}
}

func TestStickyProxySelectorPinsPerSessionKey(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
			mustParseURL(t, "http://127.0.0.1:7898"),
		},
	}
	selector := steam.NewStickyProxySelector(base)

	reqA := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")
	reqB := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-b"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	firstA, err := selector.Next(reqA)
	if err != nil {
		t.Fatalf("selector.Next(reqA) returned error: %v", err)
	}
	secondA, err := selector.Next(reqA)
	if err != nil {
		t.Fatalf("selector.Next(reqA) returned error: %v", err)
	}
	firstB, err := selector.Next(reqB)
	if err != nil {
		t.Fatalf("selector.Next(reqB) returned error: %v", err)
	}
	secondB, err := selector.Next(reqB)
	if err != nil {
		t.Fatalf("selector.Next(reqB) returned error: %v", err)
	}

	if got := firstA.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected session-a proxy: %s", got)
	}
	if got := secondA.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected repeated session-a proxy: %s", got)
	}
	if got := firstB.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected session-b proxy: %s", got)
	}
	if got := secondB.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected repeated session-b proxy: %s", got)
	}
	if base.calls.Load() != 2 {
		t.Fatalf("expected two base calls, got %d", base.calls.Load())
	}
}

func TestStickyProxySelectorFallsBackWithoutSessionKey(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
			mustParseURL(t, "http://127.0.0.1:7898"),
		},
	}
	selector := steam.NewStickyProxySelector(base)
	req := mustRequest(t, "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if got := first.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected first proxy: %s", got)
	}
	if got := second.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected second proxy: %s", got)
	}
	if base.calls.Load() != 2 {
		t.Fatalf("expected two base calls, got %d", base.calls.Load())
	}
}

func TestStickyProxySelectorTreatsBlankKeyAsUnset(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
			mustParseURL(t, "http://127.0.0.1:7898"),
		},
	}
	selector := steam.NewStickyProxySelector(base)
	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "   "), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if got := first.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected first proxy: %s", got)
	}
	if got := second.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected second proxy: %s", got)
	}
}

func TestStickyProxySelectorDoesNotCacheErrors(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
		},
		nextErrors: []error{
			errors.New("boom"),
		},
	}
	selector := steam.NewStickyProxySelector(base)
	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	if _, err := selector.Next(req); err == nil {
		t.Fatal("expected error")
	}
	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if got := second.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected recovered proxy: %s", got)
	}
	if base.calls.Load() != 2 {
		t.Fatalf("expected two base calls, got %d", base.calls.Load())
	}
}

func TestStickyProxySelectorCachesDirectConnections(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{nil},
	}
	selector := steam.NewStickyProxySelector(base)
	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if first != nil || second != nil {
		t.Fatal("expected direct connection to remain nil")
	}
	if base.calls.Load() != 1 {
		t.Fatalf("expected one base call, got %d", base.calls.Load())
	}
}

func TestStickyProxySelectorClonesReturnedURLs(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
		},
	}
	selector := steam.NewStickyProxySelector(base)
	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	first.Host = "mutated:1234"

	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	if got := second.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected cached proxy after mutation: %s", got)
	}
}

func TestStickyProxySelectorConcurrentSameSessionKey(t *testing.T) {
	t.Parallel()

	base := &recordingProxySelector{
		nextResults: []*url.URL{
			mustParseURL(t, "http://127.0.0.1:7897"),
		},
	}
	selector := steam.NewStickyProxySelector(base)

	var wg sync.WaitGroup
	results := make(chan string, 16)
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")
			proxyURL, err := selector.Next(req)
			if err != nil {
				t.Errorf("selector.Next returned error: %v", err)
				return
			}
			if proxyURL == nil {
				t.Error("expected proxy url")
				return
			}
			results <- proxyURL.String()
		}()
	}
	wg.Wait()
	close(results)

	for result := range results {
		if result != "http://127.0.0.1:7897" {
			t.Fatalf("unexpected concurrent proxy: %s", result)
		}
	}
	if base.calls.Load() != 1 {
		t.Fatalf("expected one base call, got %d", base.calls.Load())
	}
}

func TestHealthCheckedRoundRobinProxySelectorCoolsDownFailedProxy(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 1,
			Cooldown:         50 * time.Millisecond,
		},
		"http://127.0.0.1:7897",
		"http://127.0.0.1:7898",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}

	reporter := selector.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	})

	first, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("first Next returned error: %v", err)
	}
	reporter.ReportProxyResult(nil, first, http.StatusTooManyRequests, nil)

	second, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("second Next returned error: %v", err)
	}
	if got := second.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("unexpected healthy proxy: %s", got)
	}

	time.Sleep(60 * time.Millisecond)

	third, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("third Next returned error: %v", err)
	}
	if got := third.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("expected cooled proxy to recover, got %s", got)
	}
}

func TestHealthCheckedRoundRobinProxySelectorReturnsErrorWhenAllCoolingDown(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 1,
			Cooldown:         time.Second,
		},
		"http://127.0.0.1:7897",
		"http://127.0.0.1:7898",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}

	reporter := selector.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	})

	first, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("first Next returned error: %v", err)
	}
	reporter.ReportProxyResult(nil, first, http.StatusInternalServerError, nil)

	second, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("second Next returned error: %v", err)
	}
	reporter.ReportProxyResult(nil, second, http.StatusTooManyRequests, nil)

	if _, err := selector.Next(&http.Request{}); !errors.Is(err, steam.ErrAllProxiesCoolingDown) {
		t.Fatalf("expected ErrAllProxiesCoolingDown, got %v", err)
	}
}

func TestHealthCheckedRoundRobinProxySelectorResetsFailureScoreOnSuccess(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 2,
			Cooldown:         time.Second,
		},
		"http://127.0.0.1:7897",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}

	reporter := selector.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	})

	proxyURL, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	reporter.ReportProxyResult(nil, proxyURL, http.StatusInternalServerError, nil)
	reporter.ReportProxyResult(nil, proxyURL, http.StatusOK, nil)
	reporter.ReportProxyResult(nil, proxyURL, http.StatusInternalServerError, nil)

	next, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	if next == nil {
		t.Fatal("expected proxy to remain available after score reset")
	}
}

func TestHealthCheckedRoundRobinProxySelectorExposesMetricsSnapshot(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 1,
			Cooldown:         40 * time.Millisecond,
		},
		"http://user:secret@127.0.0.1:7897",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}

	metricsProvider, ok := selector.(steam.ProxyMetricsProvider)
	if !ok {
		t.Fatal("expected selector to implement ProxyMetricsProvider")
	}
	reporter := selector.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	})

	proxyURL, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	reporter.ReportProxyResult(nil, proxyURL, http.StatusTooManyRequests, nil)

	snapshot := metricsProvider.ProxyMetricsSnapshot()
	if snapshot.TotalProxies != 1 {
		t.Fatalf("unexpected total proxies: %d", snapshot.TotalProxies)
	}
	if snapshot.HealthyProxies != 0 || snapshot.CoolingProxies != 1 {
		t.Fatalf("unexpected pool status: healthy=%d cooling=%d", snapshot.HealthyProxies, snapshot.CoolingProxies)
	}
	if snapshot.GeneratedAt.IsZero() {
		t.Fatal("expected generated timestamp")
	}
	if len(snapshot.Proxies) != 1 {
		t.Fatalf("unexpected proxy metric count: %d", len(snapshot.Proxies))
	}

	metrics := snapshot.Proxies[0]
	if metrics.ProxyURL != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected proxy url: %s", metrics.ProxyURL)
	}
	if metrics.SelectionCount != 1 {
		t.Fatalf("unexpected selection count: %d", metrics.SelectionCount)
	}
	if metrics.SuccessCount != 0 {
		t.Fatalf("unexpected success count: %d", metrics.SuccessCount)
	}
	if metrics.FailureCount != 1 {
		t.Fatalf("unexpected failure count: %d", metrics.FailureCount)
	}
	if metrics.CooldownCount != 1 {
		t.Fatalf("unexpected cooldown count: %d", metrics.CooldownCount)
	}
	if metrics.FailureScore != 0 {
		t.Fatalf("expected failure score reset after cooldown, got %d", metrics.FailureScore)
	}
	if metrics.CooldownUntil.IsZero() {
		t.Fatal("expected cooldown deadline")
	}
	if metrics.LastFailureAt.IsZero() {
		t.Fatal("expected last failure timestamp")
	}
	if !metrics.LastSuccessAt.IsZero() {
		t.Fatal("did not expect success timestamp")
	}

	time.Sleep(50 * time.Millisecond)

	recovered := metricsProvider.ProxyMetricsSnapshot()
	if recovered.HealthyProxies != 1 || recovered.CoolingProxies != 0 {
		t.Fatalf("expected cooled proxy to recover in snapshot, got healthy=%d cooling=%d", recovered.HealthyProxies, recovered.CoolingProxies)
	}
}

func TestHealthCheckedRoundRobinProxySelectorSnapshotIsReadOnlyCopy(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{},
		"http://127.0.0.1:7897",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}

	metricsProvider := selector.(steam.ProxyMetricsProvider)
	first := metricsProvider.ProxyMetricsSnapshot()
	first.Proxies[0].ProxyURL = "mutated"
	first.Proxies = nil

	second := metricsProvider.ProxyMetricsSnapshot()
	if len(second.Proxies) != 1 {
		t.Fatalf("unexpected proxy metric count after mutation: %d", len(second.Proxies))
	}
	if second.Proxies[0].ProxyURL != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected proxy url after mutation: %s", second.Proxies[0].ProxyURL)
	}
}

func TestHealthCheckedRoundRobinProxySelectorTracksSuccessMetrics(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{},
		"http://127.0.0.1:7897",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}

	metricsProvider := selector.(steam.ProxyMetricsProvider)
	reporter := selector.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	})

	proxyURL, err := selector.Next(&http.Request{})
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	reporter.ReportProxyResult(nil, proxyURL, http.StatusNotFound, nil)

	snapshot := metricsProvider.ProxyMetricsSnapshot()
	metrics := snapshot.Proxies[0]
	if metrics.SelectionCount != 1 {
		t.Fatalf("unexpected selection count: %d", metrics.SelectionCount)
	}
	if metrics.SuccessCount != 1 {
		t.Fatalf("unexpected success count: %d", metrics.SuccessCount)
	}
	if metrics.FailureCount != 0 {
		t.Fatalf("unexpected failure count: %d", metrics.FailureCount)
	}
	if metrics.FailureScore != 0 {
		t.Fatalf("unexpected failure score: %d", metrics.FailureScore)
	}
	if metrics.LastSuccessAt.IsZero() {
		t.Fatal("expected last success timestamp")
	}
}

func TestHealthCheckedRoundRobinProxySelectorTreatsPublicStoreForbiddenAsFailure(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 1,
			Cooldown:         time.Second,
		},
		"http://127.0.0.1:7897",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}

	reporter := selector.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	})
	req := mustRequestWithContext(t, traffic.WithBlockDetection(steam.WithTrafficClass(context.Background(), steam.TrafficClassPublicStorePage)), "https://store.steampowered.com/app/10")

	proxyURL, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	reporter.ReportProxyResult(req, proxyURL, http.StatusForbidden, nil)

	if _, err := selector.Next(req); !errors.Is(err, steam.ErrAllProxiesCoolingDown) {
		t.Fatalf("expected ErrAllProxiesCoolingDown after public-store 403, got %v", err)
	}
}

func TestHealthCheckedRoundRobinProxySelectorIgnoresOfficialAPICode403ForCooling(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 1,
			Cooldown:         time.Second,
		},
		"http://127.0.0.1:7897",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}

	reporter := selector.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	})
	req := mustRequestWithContext(t, context.Background(), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	proxyURL, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	reporter.ReportProxyResult(req, proxyURL, http.StatusForbidden, nil)

	next, err := selector.Next(req)
	if err != nil {
		t.Fatalf("expected proxy to remain available, got %v", err)
	}
	if next == nil {
		t.Fatal("expected proxy url")
	}
}

func TestHealthCheckedRoundRobinProxySelectorRequiresBlockDetectionFlagForPublicStore403(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 1,
			Cooldown:         time.Second,
		},
		"http://127.0.0.1:7897",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}

	reporter := selector.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	})
	req := mustRequestWithContext(t, steam.WithTrafficClass(context.Background(), steam.TrafficClassPublicStorePage), "https://store.steampowered.com/app/10")

	proxyURL, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	reporter.ReportProxyResult(req, proxyURL, http.StatusForbidden, nil)

	next, err := selector.Next(req)
	if err != nil {
		t.Fatalf("expected proxy to remain available without block flag, got %v", err)
	}
	if next == nil {
		t.Fatal("expected proxy url")
	}
}

func TestStickyProxySelectorRebindsWhenHealthCheckedProxyCoolsDown(t *testing.T) {
	t.Parallel()

	base, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 1,
			Cooldown:         time.Second,
		},
		"http://127.0.0.1:7897",
		"http://127.0.0.1:7898",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}
	selector := steam.NewStickyProxySelector(base)

	reporter := selector.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	})
	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	reporter.ReportProxyResult(req, first, http.StatusTooManyRequests, nil)

	second, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	if got := second.String(); got != "http://127.0.0.1:7898" {
		t.Fatalf("expected sticky selector to rebind, got %s", got)
	}
}

func TestStickyProxySelectorForwardsHealthMetricsSnapshot(t *testing.T) {
	t.Parallel()

	base, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{},
		"http://127.0.0.1:7897",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}
	selector := steam.NewStickyProxySelector(base)

	metricsProvider, ok := selector.(steam.ProxyMetricsProvider)
	if !ok {
		t.Fatal("expected sticky selector to implement ProxyMetricsProvider")
	}

	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")
	if _, err := selector.Next(req); err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}

	snapshot := metricsProvider.ProxyMetricsSnapshot()
	if len(snapshot.Proxies) != 1 {
		t.Fatalf("unexpected proxy metric count: %d", len(snapshot.Proxies))
	}
	if snapshot.Proxies[0].SelectionCount != 1 {
		t.Fatalf("unexpected forwarded selection count: %d", snapshot.Proxies[0].SelectionCount)
	}
}

func TestStickyProxySelectorReturnsCoolingErrorWhenPoolUnavailable(t *testing.T) {
	t.Parallel()

	base, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 1,
			Cooldown:         time.Second,
		},
		"http://127.0.0.1:7897",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinProxySelector returned error: %v", err)
	}
	selector := steam.NewStickyProxySelector(base)

	reporter := selector.(interface {
		ReportProxyResult(req *http.Request, proxyURL *url.URL, statusCode int, err error)
	})
	req := mustRequestWithContext(t, steam.WithProxySessionKey(context.Background(), "session-a"), "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")

	first, err := selector.Next(req)
	if err != nil {
		t.Fatalf("selector.Next returned error: %v", err)
	}
	reporter.ReportProxyResult(req, first, http.StatusInternalServerError, nil)

	if _, err := selector.Next(req); !errors.Is(err, steam.ErrAllProxiesCoolingDown) {
		t.Fatalf("expected ErrAllProxiesCoolingDown, got %v", err)
	}
}

func TestNewRoutingProxySelector(t *testing.T) {
	t.Parallel()

	selector, err := steam.NewRoutingProxySelector(
		steam.ProxyRoute{
			Host:       "api.steampowered.com",
			PathPrefix: "/ISteamUser/",
			ProxyURL:   "http://127.0.0.1:7897",
		},
		steam.ProxyRoute{
			Host:       "steamcommunity.com",
			PathPrefix: "/openid/",
			ProxyURL:   "",
		},
	)
	if err != nil {
		t.Fatalf("NewRoutingProxySelector returned error: %v", err)
	}
	if selector == nil {
		t.Fatal("expected selector")
	}

	reqAPI := mustRequest(t, "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/")
	reqOpenID := mustRequest(t, "https://steamcommunity.com/openid/login")
	reqOther := mustRequest(t, "https://store.steampowered.com/")

	apiProxy, err := selector.Next(reqAPI)
	if err != nil {
		t.Fatalf("selector.Next(api) returned error: %v", err)
	}
	if got := apiProxy.String(); got != "http://127.0.0.1:7897" {
		t.Fatalf("unexpected api proxy: %s", got)
	}

	openIDProxy, err := selector.Next(reqOpenID)
	if err != nil {
		t.Fatalf("selector.Next(openid) returned error: %v", err)
	}
	if openIDProxy != nil {
		t.Fatal("expected direct route for openid")
	}

	otherProxy, err := selector.Next(reqOther)
	if err != nil {
		t.Fatalf("selector.Next(other) returned error: %v", err)
	}
	if otherProxy != nil {
		t.Fatal("expected direct fallback for unmatched route")
	}
}

func TestNewRoutingProxySelectorRejectsInvalidProxyURL(t *testing.T) {
	t.Parallel()

	_, err := steam.NewRoutingProxySelector(steam.ProxyRoute{
		Host:     "api.steampowered.com",
		ProxyURL: "://bad",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewHTTPClientWithProxySelector(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	selector := &countingProxySelector{}
	httpClient, err := steam.NewHTTPClientWithProxySelector(selector, 2*time.Second)
	if err != nil {
		t.Fatalf("NewHTTPClientWithProxySelector returned error: %v", err)
	}

	resp, err := httpClient.Get(server.URL)
	if err != nil {
		t.Fatalf("httpClient.Get returned error: %v", err)
	}
	_ = resp.Body.Close()

	if got := selector.calls.Load(); got == 0 {
		t.Fatal("expected proxy selector to be used by http client")
	}
}

func TestNewHTTPClientWithProxySelectorRejectsNegativeTimeout(t *testing.T) {
	t.Parallel()

	if _, err := steam.NewHTTPClientWithProxySelector(nil, -time.Second); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewHTTPClientWithProxySelectorDefaultsZeroTimeout(t *testing.T) {
	t.Parallel()

	client, err := steam.NewHTTPClientWithProxySelector(nil, 0)
	if err != nil {
		t.Fatalf("NewHTTPClientWithProxySelector returned error: %v", err)
	}
	if got := client.Timeout; got != 10*time.Second {
		t.Fatalf("unexpected timeout: %s", got)
	}
}

type countingProxySelector struct {
	calls atomic.Int32
}

func (s *countingProxySelector) Next(*http.Request) (*url.URL, error) {
	s.calls.Add(1)
	return nil, nil
}

func mustRequest(t *testing.T, rawURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("http.NewRequest returned error: %v", err)
	}
	return req
}

func mustRequestWithContext(t *testing.T, ctx context.Context, rawURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext returned error: %v", err)
	}
	return req
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}
	return parsed
}

type recordingProxySelector struct {
	calls       atomic.Int32
	mu          sync.Mutex
	nextResults []*url.URL
	nextErrors  []error
}

func (s *recordingProxySelector) Next(*http.Request) (*url.URL, error) {
	s.calls.Add(1)

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.nextErrors) > 0 {
		err := s.nextErrors[0]
		s.nextErrors = s.nextErrors[1:]
		return nil, err
	}
	if len(s.nextResults) == 0 {
		return nil, nil
	}
	result := s.nextResults[0]
	s.nextResults = s.nextResults[1:]
	if result == nil {
		return nil, nil
	}
	cloned := *result
	return &cloned, nil
}
