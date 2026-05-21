package assets

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
