package assets

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
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
