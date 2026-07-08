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
	allowed := []string{
		"https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/550/header.jpg",
		"https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/550/header.jpg",
		"https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/550/header.jpg",
		"https://shared.st.dl.eccdnx.com/store_item_assets/steam/apps/550/header.jpg",
		"https://shared.cdn.steamchina.queniuam.com/store_item_assets/steam/apps/550/header.jpg",
	}
	for _, rawURL := range allowed {
		if err := validateDirectURL(rawURL, SteamStaticURLValidator); err != nil {
			t.Fatalf("SteamStaticURLValidator rejected %s: %v", rawURL, err)
		}
	}

	rejected := []string{
		"https://example.com/header.jpg",
		"https://shared.st.dl.eccdnx.com.evil.example/header.jpg",
		"https://shared.cdn.steamchina.queniuam.com.evil.example/header.jpg",
	}
	for _, rawURL := range rejected {
		if err := validateDirectURL(rawURL, SteamStaticURLValidator); err == nil {
			t.Fatalf("SteamStaticURLValidator accepted %s", rawURL)
		}
	}
}
