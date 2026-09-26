package freeclaim

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
)

func TestResolveFreePackagesParsesPackageGroups(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/appdetails" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("appids") != "10" || query.Get("cc") != "us" || query.Get("l") != "english" {
			t.Fatalf("unexpected appdetails query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"10":{"success":true,"data":{"steam_appid":10,"name":"Demo Game","package_groups":[{"name":"default","subs":[{"packageid":100,"option_text":"Claim for free","is_free_license":true,"price_in_cents_with_discount":0},{"packageid":200,"option_text":"Totally free weekend","price_in_cents_with_discount":0},{"packageid":300,"option_text":"Paid","price_in_cents_with_discount":999},{"packageid":100,"option_text":"duplicate","is_free_license":true,"price_in_cents_with_discount":0}]}]}}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client())
	packages, err := client.ResolveFreePackages(context.Background(), 10, &ResolveFreePackagesOptions{
		CountryCode: "us",
		Language:    "english",
	})
	if err != nil {
		t.Fatalf("ResolveFreePackages returned error: %v", err)
	}
	if len(packages) != 2 {
		t.Fatalf("expected 2 free packages, got %d", len(packages))
	}
	if packages[0].PackageID != 100 || packages[0].Title != "Demo Game" {
		t.Fatalf("unexpected first package: %#v", packages[0])
	}
	if packages[1].PackageID != 200 {
		t.Fatalf("unexpected second package: %#v", packages[1])
	}
}

func TestResolveFreePackagesKeyDrift(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/appdetails" || r.URL.Query().Get("appids") != "550" {
			t.Errorf("unexpected request path or AppID")
		}
		_, _ = w.Write([]byte(`{"322070":{"success":true,"data":{"steam_appid":550,"name":"Fixture Game","package_groups":[{"subs":[{"packageid":100,"is_free_license":true,"price_in_cents_with_discount":0}]}]}}}`))
	}))
	defer server.Close()
	client := newTestClient(t, server.URL, server.Client())
	packages, err := client.ResolveFreePackages(context.Background(), 550, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(packages) != 1 || packages[0].AppID != 550 || packages[0].PackageID != 100 || packages[0].Title != "Fixture Game" {
		t.Fatalf("unexpected packages: %#v", packages)
	}
}

func TestResolveFreePackagesPreservesVerifyErrorBoundary(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"550":{"success":true,"data":{"steam_appid":620}},"322070":{"success":true,"data":{"steam_appid":550}}}`))
	}))
	defer server.Close()
	client := newTestClient(t, server.URL, server.Client())
	packages, err := client.ResolveFreePackages(context.Background(), 550, nil)
	var verifyErr *Error
	var apiErr *sdkerrors.APIError
	if !errors.As(err, &verifyErr) || verifyErr.Code != ErrorCodeVerify || verifyErr.Op != "resolve_free_packages" {
		t.Fatalf("expected existing verification boundary, got %v", err)
	}
	if !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindAPIResponse || len(packages) != 0 {
		t.Fatalf("expected wrapped identity failure without packages, got %v, %#v", err, packages)
	}
}
