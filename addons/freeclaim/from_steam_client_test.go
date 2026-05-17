package freeclaim

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	steam "github.com/gofurry/steam-go"
)

func TestNewClientFromSteamClientRejectsWithHTTPClient(t *testing.T) {
	t.Parallel()

	sdk, err := steam.NewClient()
	if err != nil {
		t.Fatalf("steam.NewClient returned error: %v", err)
	}

	_, err = NewClientFromSteamClient(sdk, WithHTTPClient(&http.Client{}))
	if err == nil {
		t.Fatal("expected configuration error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected freeclaim error, got %T", err)
	}
	if clientErr.Code != ErrorCodeConfig {
		t.Fatalf("unexpected error code: %s", clientErr.Code)
	}
}

func TestClaimPackageFromSteamClientClassifiesChallengeBlock(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dynamicstore/userdata/":
			_, _ = w.Write([]byte(`{"rgOwnedApps":[]}`))
		case "/app/10/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte("<html><body>verify you are human with g-recaptcha</body></html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}
	sdk, err := steam.NewClient(
		steam.WithStorefrontBaseURL(server.URL),
		steam.WithHTTPClient(server.Client()),
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			BlockPolicy: &steam.TrafficBlockPolicy{},
		}),
	)
	if err != nil {
		t.Fatalf("steam.NewClient returned error: %v", err)
	}
	client, err := NewClientFromSteamClient(sdk, WithStoreBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClientFromSteamClient returned error: %v", err)
	}

	result, err := client.ClaimPackage(context.Background(), NewStaticCookieJarProvider(jar), ClaimPackageRequest{
		AppID:     10,
		PackageID: 100,
	})
	if err != nil {
		t.Fatalf("ClaimPackage returned error: %v", err)
	}
	if result.Status != ClaimStatusRateLimited {
		t.Fatalf("unexpected claim result: %#v", result)
	}
}
