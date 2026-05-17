package freeclaim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
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
		_, _ = w.Write([]byte(`{"10":{"success":true,"data":{"name":"Demo Game","package_groups":[{"name":"default","subs":[{"packageid":100,"option_text":"Claim for free","is_free_license":true,"price_in_cents_with_discount":0},{"packageid":200,"option_text":"Totally free weekend","price_in_cents_with_discount":0},{"packageid":300,"option_text":"Paid","price_in_cents_with_discount":999},{"packageid":100,"option_text":"duplicate","is_free_license":true,"price_in_cents_with_discount":0}]}]}}}`))
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
