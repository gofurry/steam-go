package live_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	steam "github.com/gofurry/steam-go"
)

func TestStorefrontSmokeUsesResolvedIdentity(t *testing.T) {
	for _, tt := range []struct {
		name string
		body string
		fail int
	}{
		{"legacy", `{"550":{"success":true,"data":{"steam_appid":550}}}`, 0},
		{"key drift", `{"322070":{"success":true,"data":{"steam_appid":550}}}`, 0},
		{"invalid direct plus alternate", `{"550":{"success":false},"322070":{"success":true,"data":{"steam_appid":550}}}`, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/appdetails" || r.URL.Query().Get("appids") != "550" {
					t.Error("unexpected storefront request")
				}
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			client, err := steam.NewClient(steam.WithStorefrontBaseURL(server.URL), steam.WithRetry(0))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			report := newLiveSmokeReport("direct")
			// This exercises the same helper as opt-in live smoke without Steam traffic.
			runLiveSmokeCheck(&report, "webstorefront", func(ctx context.Context) error {
				return checkStorefrontAppDetails(ctx, client.Web.Storefront)
			})
			if report.Summary.Fail != tt.fail {
				t.Fatalf("summary=%#v", report.Summary)
			}
		})
	}
}
