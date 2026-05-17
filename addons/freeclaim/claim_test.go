package freeclaim

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/gofurry/steam-go/addons/websession"
)

func TestClaimPackageUsesProviderAndOwnershipFallback(t *testing.T) {
	t.Parallel()

	var ownershipChecks atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dynamicstore/userdata/":
			checkNumber := ownershipChecks.Add(1)
			if checkNumber == 1 {
				_, _ = w.Write([]byte(`{"rgOwnedApps":[]}`))
				return
			}
			_, _ = w.Write([]byte(`{"rgOwnedApps":[10]}`))
		case "/app/10/":
			_, _ = w.Write([]byte(`<html><body><form name="add_to_cart_100"><input type="hidden" name="snr" value="1_abc"/><input type="hidden" name="originating_snr" value="1_origin"/></form></body></html>`))
		case "/checkout/addfreelicense":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm returned error: %v", err)
			}
			if got := r.Form.Get("action"); got != "add_to_cart" {
				t.Fatalf("unexpected action: %q", got)
			}
			if got := r.Form.Get("subid"); got != "100" {
				t.Fatalf("unexpected subid: %q", got)
			}
			if got := r.Form.Get("sessionid"); got != "store-session" {
				t.Fatalf("unexpected sessionid: %q", got)
			}
			if got := r.Header.Get("Origin"); got != server.URL {
				t.Fatalf("unexpected origin: %q", got)
			}
			if got := r.Header.Get("Referer"); got != server.URL+"/app/10/" {
				t.Fatalf("unexpected referer: %q", got)
			}
			if got := r.Header.Get("X-Requested-With"); got != "XMLHttpRequest" {
				t.Fatalf("unexpected x-requested-with: %q", got)
			}
			_, _ = w.Write([]byte(`processing`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}
	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}
	jar.SetCookies(parsedURL, []*http.Cookie{{
		Name:  "sessionid",
		Value: "store-session",
		Path:  "/",
	}})

	client := newTestClient(t, server.URL, server.Client())
	result, err := client.ClaimPackage(context.Background(), &websession.WebCookieResult{Jar: jar}, ClaimPackageRequest{
		AppID:     10,
		PackageID: 100,
	})
	if err != nil {
		t.Fatalf("ClaimPackage returned error: %v", err)
	}
	if result.Status != ClaimStatusClaimed || !result.Owned {
		t.Fatalf("unexpected claim result: %#v", result)
	}
}

func TestClaimPackageReturnsAlreadyOwnedBeforePosting(t *testing.T) {
	t.Parallel()

	var claimPosted atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dynamicstore/userdata/":
			_, _ = w.Write([]byte(`{"rgOwnedApps":[10]}`))
		case "/checkout/addfreelicense":
			claimPosted.Store(true)
			_, _ = w.Write([]byte(`should not be called`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}
	client := newTestClient(t, server.URL, server.Client())

	result, err := client.ClaimPackage(context.Background(), NewStaticCookieJarProvider(jar), ClaimPackageRequest{
		AppID:     10,
		PackageID: 100,
	})
	if err != nil {
		t.Fatalf("ClaimPackage returned error: %v", err)
	}
	if result.Status != ClaimStatusAlreadyOwned || !result.Owned {
		t.Fatalf("unexpected claim result: %#v", result)
	}
	if claimPosted.Load() {
		t.Fatal("expected claim endpoint not to be called")
	}
}

func TestClaimPackageClassifiesRateLimit(t *testing.T) {
	t.Parallel()

	var ownershipChecks atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dynamicstore/userdata/":
			ownershipChecks.Add(1)
			_, _ = w.Write([]byte(`{"rgOwnedApps":[]}`))
		case "/app/10/":
			_, _ = w.Write([]byte(`<html><body><form name="add_to_cart_100"><input type="hidden" name="sessionid" value="store-session"/></form></body></html>`))
		case "/checkout/addfreelicense":
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`slow down`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}
	client := newTestClient(t, server.URL, server.Client())

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
