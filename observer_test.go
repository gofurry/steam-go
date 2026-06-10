package steam_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	steam "github.com/gofurry/steam-go"
)

func TestRequestObserverReceivesSanitizedSuccessEvent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	events := &eventRecorder{}
	client, err := steam.NewClient(steam.WithRequestObserver(events))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/ok?key=SECRET&steamids=1", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}
	_, err = client.DoRawHTTPRequest(context.Background(), req, &steam.RawHTTPRequestOptions{TrafficClass: steam.TrafficClassMarketWeb})
	if err != nil {
		t.Fatalf("DoRawHTTPRequest returned error: %v", err)
	}

	event := events.single(t)
	if event.TrafficClass != steam.TrafficClassMarketWeb {
		t.Fatalf("unexpected traffic class: %q", event.TrafficClass)
	}
	if event.Path != "/ok" {
		t.Fatalf("observer path should exclude query, got %q", event.Path)
	}
	if event.StatusCode != http.StatusOK || event.ErrorKind != "" || event.Attempts != 1 {
		t.Fatalf("unexpected event: %#v", event)
	}
}

func TestRequestObserverReceivesHTTPStatusErrorKind(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer server.Close()

	events := &eventRecorder{}
	client, err := steam.NewClient(
		steam.WithCommunityBaseURL(server.URL),
		steam.WithRequestObserver(events),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.Web.Market.GetPriceOverview(context.Background(), 440, "Key", nil)
	if err == nil {
		t.Fatal("expected HTTP status error")
	}
	event := events.single(t)
	if event.ErrorKind != string(steam.ErrorKindHTTPStatus) || event.StatusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected error event: %#v", event)
	}
	if event.Path != "/market/priceoverview" {
		t.Fatalf("unexpected path: %q", event.Path)
	}
}

func TestRequestObserverReceivesRetryAndCacheEvents(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		request := requests.Add(1)
		w.Header().Set("Cache-Control", "max-age=60")
		if request == 1 {
			http.Error(w, "retry", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	events := &eventRecorder{}
	client, err := steam.NewClient(
		steam.WithRequestObserver(events),
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			Retry: &steam.TrafficRetryPolicy{
				Retry:   1,
				Backoff: steam.RetryBackoffConfig{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond},
			},
			Cache: &steam.TrafficCachePolicy{TTL: time.Minute},
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	makeReq := func() *http.Request {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/cache", nil)
		if err != nil {
			t.Fatalf("NewRequestWithContext returned error: %v", err)
		}
		return req
	}
	if _, err := client.DoRawHTTPRequest(context.Background(), makeReq(), &steam.RawHTTPRequestOptions{TrafficClass: steam.TrafficClassPublicStorePage}); err != nil {
		t.Fatalf("first DoRawHTTPRequest returned error: %v", err)
	}
	if _, err := client.DoRawHTTPRequest(context.Background(), makeReq(), &steam.RawHTTPRequestOptions{TrafficClass: steam.TrafficClassPublicStorePage}); err != nil {
		t.Fatalf("second DoRawHTTPRequest returned error: %v", err)
	}

	recorded := events.all()
	if len(recorded) != 2 {
		t.Fatalf("expected 2 events, got %#v", recorded)
	}
	if recorded[0].Attempts != 2 || recorded[0].CacheHit {
		t.Fatalf("expected first event to report retry attempts, got %#v", recorded[0])
	}
	if !recorded[1].CacheHit || recorded[1].Attempts != 1 {
		t.Fatalf("expected second event to report cache hit, got %#v", recorded[1])
	}
}

func TestRequestObserverReceivesConditionalCacheRefreshEvent(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requests.Add(1) {
		case 1:
			w.Header().Set("ETag", `"etag-a"`)
			_, _ = w.Write([]byte("cached"))
		case 2:
			if got := r.Header.Get("If-None-Match"); got != `"etag-a"` {
				t.Fatalf("unexpected If-None-Match: %q", got)
			}
			w.WriteHeader(http.StatusNotModified)
		default:
			t.Fatalf("unexpected extra request")
		}
	}))
	defer server.Close()

	events := &eventRecorder{}
	client, err := steam.NewClient(
		steam.WithRequestObserver(events),
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			Cache: &steam.TrafficCachePolicy{TTL: 10 * time.Millisecond},
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	makeReq := func() *http.Request {
		req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/cached", nil)
		if reqErr != nil {
			t.Fatalf("NewRequestWithContext returned error: %v", reqErr)
		}
		return req
	}
	if _, err := client.DoRawHTTPRequest(context.Background(), makeReq(), &steam.RawHTTPRequestOptions{TrafficClass: steam.TrafficClassPublicStorePage}); err != nil {
		t.Fatalf("first DoRawHTTPRequest returned error: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	if _, err := client.DoRawHTTPRequest(context.Background(), makeReq(), &steam.RawHTTPRequestOptions{TrafficClass: steam.TrafficClassPublicStorePage}); err != nil {
		t.Fatalf("second DoRawHTTPRequest returned error: %v", err)
	}

	recorded := events.all()
	if len(recorded) != 2 {
		t.Fatalf("expected 2 events, got %#v", recorded)
	}
	refresh := recorded[1]
	if refresh.StatusCode != http.StatusNotModified || !refresh.CacheHit || !refresh.ConditionalHit {
		t.Fatalf("expected conditional cache refresh event, got %#v", refresh)
	}
}

func TestRequestObserverReceivesBlockDetectedEvent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>captcha access denied</body></html>"))
	}))
	defer server.Close()

	events := &eventRecorder{}
	client, err := steam.NewClient(
		steam.WithRequestObserver(events),
		steam.WithTrafficPolicy(steam.TrafficClassCommunityWeb, steam.TrafficPolicy{
			BlockPolicy: &steam.TrafficBlockPolicy{},
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/challenge", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}
	result, err := client.DoRawHTTPRequest(context.Background(), req, &steam.RawHTTPRequestOptions{TrafficClass: steam.TrafficClassCommunityWeb})
	if err != nil {
		t.Fatalf("DoRawHTTPRequest returned error: %v", err)
	}
	if result.Block == nil {
		t.Fatal("expected raw block metadata")
	}

	event := events.single(t)
	if !event.BlockDetected {
		t.Fatalf("expected block detected event, got %#v", event)
	}
}

type eventRecorder struct {
	mu     sync.Mutex
	events []steam.RequestEvent
}

func (r *eventRecorder) ObserveRequest(event steam.RequestEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
}

func (r *eventRecorder) single(t *testing.T) steam.RequestEvent {
	t.Helper()
	events := r.all()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %#v", events)
	}
	return events[0]
}

func (r *eventRecorder) all() []steam.RequestEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]steam.RequestEvent(nil), r.events...)
}
