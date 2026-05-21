package assets

import (
	"context"
	"net/http"
	"testing"
)

func TestURLValidatorRejectsDirectURLs(t *testing.T) {
	server := newAssetTestServer(t)

	results, err := ReadURLsWithOptions(context.Background(), ReadOptions{
		URLValidator: AllowHosts("example.com"),
	}, server.URL+"/header.jpg")
	if err == nil {
		t.Fatal("ReadURLsWithOptions returned nil error")
	}
	if len(results) != 1 || results[0].URL != server.URL+"/header.jpg" || results[0].Error == "" {
		t.Fatalf("results = %#v", results)
	}

	downloadResults, err := DownloadURLsWithOptions(context.Background(), DownloadOptions{
		Dir:          t.TempDir(),
		URLValidator: AllowHosts("example.com"),
	}, server.URL+"/header.jpg")
	if err == nil {
		t.Fatal("DownloadURLsWithOptions returned nil error")
	}
	if len(downloadResults) != 1 || downloadResults[0].URL != server.URL+"/header.jpg" || downloadResults[0].Error == "" {
		t.Fatalf("download results = %#v", downloadResults)
	}
}

func TestVerifyURLsWithOptionsValidator(t *testing.T) {
	server := newAssetTestServer(t)

	_, err := VerifyURLsWithOptions(context.Background(), VerifyOptions{
		URLValidator: AllowHosts("example.com"),
	}, server.URL+"/header.jpg")
	if err == nil {
		t.Fatal("VerifyURLsWithOptions returned nil error")
	}

	results, err := VerifyURLsWithOptions(context.Background(), VerifyOptions{
		HTTPClient:   http.DefaultClient,
		URLValidator: AllowHosts(server.Listener.Addr().String()),
	}, server.URL+"/header.jpg")
	if err != nil {
		t.Fatalf("VerifyURLsWithOptions returned error: %v", err)
	}
	if len(results) != 1 || !results[0].Exists {
		t.Fatalf("results = %#v", results)
	}
}

func TestSteamStaticURLValidator(t *testing.T) {
	if err := validateDirectURL("https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/550/header.jpg", SteamStaticURLValidator); err != nil {
		t.Fatalf("SteamStaticURLValidator rejected steamstatic URL: %v", err)
	}
	if err := validateDirectURL("https://example.com/header.jpg", SteamStaticURLValidator); err == nil {
		t.Fatal("SteamStaticURLValidator accepted non-steamstatic URL")
	}
}
