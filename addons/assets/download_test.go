package assets

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownloadURLs(t *testing.T) {
	server := newAssetTestServer(t)
	dir := t.TempDir()

	results, err := DownloadURLs(context.Background(), dir, server.URL+"/header.jpg")
	if err != nil {
		t.Fatalf("DownloadURLs returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("DownloadURLs returned %d results, want 1", len(results))
	}
	if results[0].BytesWritten != int64(len("header-body")) {
		t.Fatalf("BytesWritten = %d", results[0].BytesWritten)
	}
	body, err := os.ReadFile(filepath.Join(dir, "header.jpg"))
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(body) != "header-body" {
		t.Fatalf("downloaded body = %q", body)
	}
}

func TestDownloadURLsContinuesAfterErrors(t *testing.T) {
	server := newAssetTestServer(t)
	dir := t.TempDir()

	results, err := DownloadURLs(
		context.Background(),
		dir,
		server.URL+"/missing.jpg",
		server.URL+"/header.jpg",
	)
	if err == nil {
		t.Fatal("DownloadURLs returned nil error")
	}
	if len(results) != 2 {
		t.Fatalf("DownloadURLs returned %d results, want 2", len(results))
	}
	if results[0].Error == "" || results[0].StatusCode != http.StatusNotFound {
		t.Fatalf("missing result = %#v", results[0])
	}
	if results[1].Error != "" || results[1].BytesWritten == 0 {
		t.Fatalf("success result = %#v", results[1])
	}
	assertFile(t, filepath.Join(dir, "header.jpg"), "header-body")
}

func TestDownloadURLsUniquifiesDuplicateFilenames(t *testing.T) {
	server := newAssetTestServer(t)
	dir := t.TempDir()

	results, err := DownloadURLsWithOptions(context.Background(), DownloadOptions{
		Dir:         dir,
		Concurrency: 2,
	}, server.URL+"/same/header.jpg", server.URL+"/other/header.jpg")
	if err != nil {
		t.Fatalf("DownloadURLsWithOptions returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if results[0].Path == results[1].Path {
		t.Fatalf("duplicate paths were not uniquified: %#v", results)
	}
	assertFile(t, filepath.Join(dir, "header.jpg"), "same-header")
	assertFile(t, filepath.Join(dir, "header_2.jpg"), "other-header")
}

func TestDownloadURLsKeepsURLOnTransportError(t *testing.T) {
	rawURL := "https://example.com/header.jpg"
	results, err := DownloadURLsWithOptions(context.Background(), DownloadOptions{
		Dir:        t.TempDir(),
		HTTPClient: &http.Client{Transport: failingTransport{}},
	}, rawURL)
	if err == nil {
		t.Fatal("DownloadURLsWithOptions returned nil error")
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].URL != rawURL || results[0].Status != DownloadStatusFailed || results[0].Error == "" {
		t.Fatalf("result = %#v", results[0])
	}
}

func TestDownloadAppAssetsStoreModes(t *testing.T) {
	client := &http.Client{Transport: staticAssetTransport{}}

	flatDir := t.TempDir()
	results, err := DownloadAppAssets(context.Background(), DownloadAppOptions{
		Dir:        flatDir,
		Kinds:      []Kind{KindHeader, KindLibraryHero},
		HTTPClient: client,
	}, 550, 107100)
	if err != nil {
		t.Fatalf("DownloadAppAssets flat returned error: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("flat results = %d, want 4", len(results))
	}
	assertFile(t, filepath.Join(flatDir, "550_header.jpg"), "asset:header.jpg")
	assertFile(t, filepath.Join(flatDir, "107100_library_hero.jpg"), "asset:library_hero.jpg")

	byAppDir := t.TempDir()
	results, err = DownloadAppAssets(context.Background(), DownloadAppOptions{
		Dir:        byAppDir,
		Kinds:      []Kind{KindHeaderLocalized},
		Language:   "schinese",
		Mode:       StoreByAppID,
		HTTPClient: client,
	}, 550)
	if err != nil {
		t.Fatalf("DownloadAppAssets by app returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("by app results = %d, want 1", len(results))
	}
	assertFile(t, filepath.Join(byAppDir, "550", "header_schinese.jpg"), "asset:header_schinese.jpg")
}

func TestDownloadAppAssetsContinuesAfterErrors(t *testing.T) {
	client := &http.Client{Transport: selectiveAssetTransport{}}
	dir := t.TempDir()

	results, err := DownloadAppAssets(context.Background(), DownloadAppOptions{
		Dir:        dir,
		Kinds:      []Kind{KindHeaderLocalized, KindHeader, KindLibraryHero},
		Language:   "schinese",
		Mode:       StoreByAppID,
		HTTPClient: client,
	}, 550)
	if err == nil {
		t.Fatal("DownloadAppAssets returned nil error")
	}
	if len(results) != 3 {
		t.Fatalf("results = %d, want 3", len(results))
	}
	if results[0].Error == "" || results[0].StatusCode != http.StatusNotFound {
		t.Fatalf("localized result = %#v", results[0])
	}
	if results[0].Status != DownloadStatusFailed {
		t.Fatalf("localized status = %q", results[0].Status)
	}
	if results[1].Error != "" || results[2].Error != "" {
		t.Fatalf("successful results have errors: %#v", results)
	}
	if results[1].Status != DownloadStatusDownloaded || results[2].Status != DownloadStatusDownloaded {
		t.Fatalf("successful statuses = %#v", results)
	}
	assertFile(t, filepath.Join(dir, "550", "header.jpg"), "asset:header.jpg")
	assertFile(t, filepath.Join(dir, "550", "library_hero.jpg"), "asset:library_hero.jpg")
}

func TestDownloadAppAssetsSkipExistingAndFilenameStyle(t *testing.T) {
	client := &http.Client{Transport: staticAssetTransport{}}
	dir := t.TempDir()
	path := filepath.Join(dir, "550", "library_hero.jpg")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := DownloadAppAssets(context.Background(), DownloadAppOptions{
		Dir:           dir,
		Kinds:         []Kind{KindLibraryHero, KindHeader},
		Mode:          StoreByAppID,
		HTTPClient:    client,
		Overwrite:     OverwriteNever,
		FilenameStyle: FilenameKind,
		Concurrency:   2,
	}, 550)
	if err != nil {
		t.Fatalf("DownloadAppAssets returned error: %v", err)
	}
	if results[0].Status != DownloadStatusSkipped {
		t.Fatalf("first status = %q, want skipped", results[0].Status)
	}
	if results[1].Status != DownloadStatusDownloaded {
		t.Fatalf("second status = %q, want downloaded", results[1].Status)
	}
	assertFile(t, path, "existing")
	assertFile(t, filepath.Join(dir, "550", "header.jpg"), "asset:header.jpg")
}

func TestDownloadAppAssetsRejectsUnsafeLocalizedLanguage(t *testing.T) {
	client := &http.Client{Transport: staticAssetTransport{}}
	dir := t.TempDir()

	_, err := DownloadAppAssets(context.Background(), DownloadAppOptions{
		Dir:        dir,
		Language:   "../bad",
		HTTPClient: client,
	}, 550)
	if err == nil {
		t.Fatal("DownloadAppAssets returned nil error")
	}
}

func TestDownloadURLsCanceledBeforeEnqueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := DownloadURLsWithOptions(ctx, DownloadOptions{
		Dir: t.TempDir(),
	}, "https://example.com/a.jpg", "https://example.com/b.jpg")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("DownloadURLsWithOptions error = %v, want context canceled", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	for _, result := range results {
		if result.Status != DownloadStatusFailed || !strings.Contains(result.Error, context.Canceled.Error()) {
			t.Fatalf("expected canceled failed result, got %#v", result)
		}
	}
}

func TestDownloadURLsStopsEnqueueAfterContextCancellation(t *testing.T) {
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
		_, err := DownloadURLsWithOptions(ctx, DownloadOptions{
			Dir:         t.TempDir(),
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
			t.Fatalf("DownloadURLsWithOptions error = %v, want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("DownloadURLsWithOptions did not return after cancellation")
	}
	if got := transport.calls.Load(); got > 2 {
		t.Fatalf("expected enqueue to stop promptly, got %d transport calls", got)
	}
}

type staticAssetTransport struct{}

func (staticAssetTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	name := filepath.Base(req.URL.Path)
	body := "asset:" + name
	return stringResponse(req, http.StatusOK, "image/test", body), nil
}

type selectiveAssetTransport struct{}

func (selectiveAssetTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	name := filepath.Base(req.URL.Path)
	if name == "header_schinese.jpg" {
		return stringResponse(req, http.StatusNotFound, "text/plain", "missing"), nil
	}
	body := "asset:" + name
	return stringResponse(req, http.StatusOK, "image/test", body), nil
}

type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("transport failed")
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if strings.TrimSpace(string(body)) != want {
		t.Fatalf("%s = %q, want %q", path, body, want)
	}
}
