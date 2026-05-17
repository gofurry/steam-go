package freeclaim

import (
	"net/http"
	"testing"

	steam "github.com/gofurry/steam-go"
)

func newTestClient(t *testing.T, baseURL string, httpClient *http.Client) *Client {
	t.Helper()

	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	sdk, err := steam.NewClient(
		steam.WithStorefrontBaseURL(baseURL),
		steam.WithHTTPClient(httpClient),
	)
	if err != nil {
		t.Fatalf("steam.NewClient returned error: %v", err)
	}

	client, err := NewClient(
		sdk.Web.Storefront,
		WithStoreBaseURL(baseURL),
		WithHTTPClient(httpClient),
	)
	if err != nil {
		t.Fatalf("freeclaim.NewClient returned error: %v", err)
	}
	return client
}

func newTestClientFromSteamClient(t *testing.T, baseURL string, httpClient *http.Client) *Client {
	t.Helper()

	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	sdk, err := steam.NewClient(
		steam.WithStorefrontBaseURL(baseURL),
		steam.WithHTTPClient(httpClient),
	)
	if err != nil {
		t.Fatalf("steam.NewClient returned error: %v", err)
	}

	client, err := NewClientFromSteamClient(
		sdk,
		WithStoreBaseURL(baseURL),
	)
	if err != nil {
		t.Fatalf("freeclaim.NewClientFromSteamClient returned error: %v", err)
	}
	return client
}
