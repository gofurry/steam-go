package assets

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestReadURLs(t *testing.T) {
	server := newAssetTestServer(t)

	results, err := ReadURLs(context.Background(), server.URL+"/header.jpg")
	if err != nil {
		t.Fatalf("ReadURLs returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("ReadURLs returned %d results, want 1", len(results))
	}
	if got := string(results[0].Data); got != "header-body" {
		t.Fatalf("Data = %q", got)
	}
	if results[0].BytesRead != int64(len("header-body")) || results[0].ContentType != "image/jpeg" {
		t.Fatalf("result = %#v", results[0])
	}
}

func TestReadURLsContinuesAfterErrors(t *testing.T) {
	server := newAssetTestServer(t)

	results, err := ReadURLs(
		context.Background(),
		server.URL+"/missing.jpg",
		server.URL+"/header.jpg",
	)
	if err == nil {
		t.Fatal("ReadURLs returned nil error")
	}
	if len(results) != 2 {
		t.Fatalf("ReadURLs returned %d results, want 2", len(results))
	}
	if results[0].Error == "" || results[0].StatusCode != http.StatusNotFound {
		t.Fatalf("missing result = %#v", results[0])
	}
	if results[1].Error != "" || string(results[1].Data) != "header-body" {
		t.Fatalf("success result = %#v", results[1])
	}
}

func TestReadURLsMaxBytes(t *testing.T) {
	server := newAssetTestServer(t)

	results, err := ReadURLsWithOptions(context.Background(), ReadOptions{
		MaxBytes: 4,
	}, server.URL+"/header.jpg")
	if err == nil {
		t.Fatal("ReadURLsWithOptions returned nil error")
	}
	if len(results) != 1 || results[0].Error == "" || len(results[0].Data) != 0 {
		t.Fatalf("max bytes result = %#v", results)
	}
}

func TestReadEachURLs(t *testing.T) {
	server := newAssetTestServer(t)

	var got []string
	err := ReadEachURLs(context.Background(), ReadOptions{}, func(result ReadResult) error {
		if result.Error != "" {
			return nil
		}
		got = append(got, string(result.Data))
		return nil
	}, server.URL+"/header.jpg", server.URL+"/missing.jpg")
	if err == nil {
		t.Fatal("ReadEachURLs returned nil error")
	}
	if len(got) != 1 || got[0] != "header-body" {
		t.Fatalf("handler data = %#v", got)
	}
}

func TestReadEachURLsReturnsHandlerErrors(t *testing.T) {
	server := newAssetTestServer(t)

	err := ReadEachURLs(context.Background(), ReadOptions{}, func(result ReadResult) error {
		return errors.New("handler failed")
	}, server.URL+"/header.jpg")
	if !containsError(err, "handler failed") {
		t.Fatalf("ReadEachURLs error = %v", err)
	}
}

func TestReadAppAssets(t *testing.T) {
	client := &http.Client{Transport: staticAssetTransport{}}

	results, err := ReadAppAssets(context.Background(), ReadAppOptions{
		Kinds:       []Kind{KindHeader, KindLibraryHero},
		HTTPClient:  client,
		Concurrency: 2,
	}, 550)
	if err != nil {
		t.Fatalf("ReadAppAssets returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if results[0].AppID != 550 || results[0].Kind != KindHeader || string(results[0].Data) != "asset:header.jpg" {
		t.Fatalf("first result = %#v", results[0])
	}
	if results[1].AppID != 550 || results[1].Kind != KindLibraryHero || string(results[1].Data) != "asset:library_hero.jpg" {
		t.Fatalf("second result = %#v", results[1])
	}
}

func containsError(err error, text string) bool {
	return err != nil && strings.Contains(err.Error(), text)
}

func TestReadStoreMedia(t *testing.T) {
	mediaServer := newStoreMediaAssetServer(t)
	service := newStoreMediaTestService(t, storeMediaAppDetails(mediaServer.URL))

	results, err := ReadStoreMedia(context.Background(), service, ReadStoreMediaOptions{
		Kinds:       []Kind{KindMovieDASHH264},
		Concurrency: 2,
	}, 550)
	if err != nil {
		t.Fatalf("ReadStoreMedia returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if results[0].Name != "Trailer One" || string(results[0].Data) != "media:h264_5952.mpd" {
		t.Fatalf("first result = %#v", results[0])
	}
	if filepath.Ext(results[1].URL) != ".mpd" {
		t.Fatalf("second URL = %q", results[1].URL)
	}
}

func TestReadURLsCanceledBeforeEnqueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := ReadURLsWithOptions(ctx, ReadOptions{}, "https://example.com/a.jpg", "https://example.com/b.jpg")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ReadURLsWithOptions error = %v, want context canceled", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	for _, result := range results {
		if !strings.Contains(result.Error, context.Canceled.Error()) {
			t.Fatalf("expected canceled result, got %#v", result)
		}
	}
}

func TestReadURLsStopsEnqueueAfterContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	transport := &cancelAwareTransport{started: make(chan struct{}, 1)}
	client := &http.Client{Transport: transport}
	urls := make([]string, 64)
	for i := range urls {
		urls[i] = "https://example.com/asset.jpg"
	}

	done := make(chan error, 1)
	go func() {
		_, err := ReadURLsWithOptions(ctx, ReadOptions{
			HTTPClient:  client,
			Concurrency: 1,
		}, urls...)
		done <- err
	}()

	select {
	case <-transport.started:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("first request was not started")
	}

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ReadURLsWithOptions error = %v, want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ReadURLsWithOptions did not return after cancellation")
	}
	if got := transport.calls.Load(); got > 2 {
		t.Fatalf("expected enqueue to stop promptly, got %d transport calls", got)
	}
}

func TestReadEachURLsCanceledBeforeEnqueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var handled atomic.Int32
	err := ReadEachURLs(ctx, ReadOptions{}, func(ReadResult) error {
		handled.Add(1)
		return nil
	}, "https://example.com/a.jpg")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ReadEachURLs error = %v, want context canceled", err)
	}
	if got := handled.Load(); got != 0 {
		t.Fatalf("expected no handled results, got %d", got)
	}
}

type cancelAwareTransport struct {
	started chan struct{}
	calls   atomic.Int32
}

func (t *cancelAwareTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.calls.Add(1)
	select {
	case t.started <- struct{}{}:
	default:
	}
	<-req.Context().Done()
	return nil, req.Context().Err()
}
