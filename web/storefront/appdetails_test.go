package storefront

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
)

func TestResolveAppDetails(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		appID uint32
		body  string
		key   string
		kind  sdkerrors.Kind
	}{
		{"zero AppID", 0, `null`, "", sdkerrors.KindRequestBuild},
		{"nil envelope", 550, `null`, "", sdkerrors.KindAPIResponse},
		{"empty envelope", 550, `{}`, "", sdkerrors.KindAPIResponse},
		{"legacy", 550, `{"550":{"success":true,"data":{"steam_appid":550}}}`, "550", ""},
		{"drift", 550, `{"322070":{"success":true,"data":{"steam_appid":550}}}`, "322070", ""},
		{"opaque key preserved", 550, `{" opaque:key ":{"success":true,"data":{"steam_appid":550}}}`, " opaque:key ", ""},
		{"empty key preserved", 550, `{"":{"success":true,"data":{"steam_appid":550}}}`, "", ""},
		{"unrelated entries", 550, `{"323180":{"success":true,"data":{"steam_appid":620}},"322070":{"success":true,"data":{"steam_appid":550}},"999999":{"success":false,"data":{"steam_appid":999999}}}`, "322070", ""},
		{"no match", 550, `{"620":{"success":true,"data":{"steam_appid":620}}}`, "", sdkerrors.KindAPIResponse},
		{"single entry without identity", 550, `{"322070":{"success":true,"data":{"name":"not identity"}}}`, "", sdkerrors.KindAPIResponse},
		{"direct conflict", 550, `{"550":{"success":true,"data":{"steam_appid":620}}}`, "", sdkerrors.KindAPIResponse},
		{"direct missing identity", 550, `{"550":{"success":true,"data":{}}}`, "", sdkerrors.KindAPIResponse},
		{"direct zero identity", 550, `{"550":{"success":true,"data":{"steam_appid":0}}}`, "", sdkerrors.KindAPIResponse},
		{"direct missing data", 550, `{"550":{"success":true}}`, "", sdkerrors.KindAPIResponse},
		{"direct failure", 550, `{"550":{"success":false,"data":{"steam_appid":550}}}`, "", sdkerrors.KindAPIResponse},
		{"direct failure without data", 550, `{"550":{"success":false}}`, "", sdkerrors.KindAPIResponse},
		{"conflict plus alternate", 550, `{"550":{"success":true,"data":{"steam_appid":620}},"322070":{"success":true,"data":{"steam_appid":550}}}`, "", sdkerrors.KindAPIResponse},
		{"missing identity plus alternate", 550, `{"550":{"success":true,"data":{}},"322070":{"success":true,"data":{"steam_appid":550}}}`, "", sdkerrors.KindAPIResponse},
		{"failure plus alternate", 550, `{"550":{"success":false},"322070":{"success":true,"data":{"steam_appid":550}}}`, "", sdkerrors.KindAPIResponse},
		{"duplicate identity", 550, `{"111":{"success":true,"data":{"steam_appid":550}},"222":{"success":true,"data":{"steam_appid":550}}}`, "", sdkerrors.KindAPIResponse},
		{"valid direct plus duplicate", 550, `{"550":{"success":true,"data":{"steam_appid":550}},"322070":{"success":true,"data":{"steam_appid":550}}}`, "", sdkerrors.KindAPIResponse},
		{"duplicate includes failure", 550, `{"111":{"success":true,"data":{"steam_appid":550}},"222":{"success":false,"data":{"steam_appid":550}}}`, "", sdkerrors.KindAPIResponse},
		{"sole identity failure", 550, `{"322070":{"success":false,"data":{"steam_appid":550}}}`, "", sdkerrors.KindAPIResponse},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var envelope AppDetailsEnvelope
			if err := json.Unmarshal([]byte(tt.body), &envelope); err != nil {
				t.Fatal(err)
			}
			before, err := json.Marshal(envelope)
			if err != nil {
				t.Fatal(err)
			}
			match, err := ResolveAppDetails(envelope, tt.appID)
			if tt.kind != "" {
				var apiErr *sdkerrors.APIError
				if !errors.As(err, &apiErr) || apiErr.Kind != tt.kind {
					t.Fatalf("error = %v, want %s", err, tt.kind)
				}
				if apiErr.StatusCode != 0 || apiErr.Body != nil {
					t.Fatal("local resolver fabricated an HTTP response")
				}
				if !reflect.DeepEqual(match, AppDetailsMatch{}) {
					t.Fatalf("failed resolution returned a usable match: %#v", match)
				}
				// Map iteration must not affect ambiguous-response diagnostics.
				for range 8 {
					_, again := ResolveAppDetails(envelope, tt.appID)
					if again == nil || again.Error() != err.Error() {
						t.Fatal("non-deterministic resolution error")
					}
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if match.RequestedAppID != tt.appID || match.ResponseKey != tt.key || !match.Result.Success || match.Result.Data.SteamAppID != tt.appID {
					t.Fatalf("unexpected match: %#v", match)
				}
				if !reflect.DeepEqual(match.Result, envelope[tt.key]) {
					t.Fatal("result was rewritten")
				}
			}
			after, err := json.Marshal(envelope)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("resolver modified the upstream envelope")
			}
		})
	}
}

func TestAppDetailsKeyDriftFidelity(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "web", "storefront", "GetAppDetails", "app_550_key_drift.json"))
	if err != nil {
		t.Fatal(err)
	}
	store := &recordingTransport{responseBody: string(body)}
	official := &recordingTransport{}
	service := newTestService(t, official, store, 4096)
	opts := &GetAppDetailsOptions{CountryCode: "US", Language: "english"}
	raw, err := service.GetAppDetailsRaw(context.Background(), 550, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, body) {
		t.Fatal("Raw changed upstream bytes")
	}
	envelope, err := service.GetAppDetails(context.Background(), 550, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(envelope) != 1 || envelope["322070"].Data.SteamAppID != 550 {
		t.Fatalf("unexpected upstream envelope: %#v", envelope)
	}
	if _, ok := envelope["550"]; ok {
		t.Fatal("typed decode manufactured a requested-AppID key")
	}
	match, err := service.GetResolvedAppDetails(context.Background(), 550, opts)
	if err != nil {
		t.Fatal(err)
	}
	if match.ResponseKey != "322070" || match.RequestedAppID != 550 || !reflect.DeepEqual(match.Result, envelope["322070"]) {
		t.Fatalf("unexpected resolved result: %#v", match)
	}
	batch, err := service.GetAppDetailsBatch(context.Background(), []uint32{550}, nil)
	if err != nil || len(batch) != 1 || batch[0].Err != nil {
		t.Fatalf("batch = %#v, err = %v", batch, err)
	}
	if batch[0].AppID != 550 || !reflect.DeepEqual(batch[0].Response, envelope) {
		t.Fatal("batch changed the upstream envelope")
	}
	if _, err := ResolveAppDetails(batch[0].Response, batch[0].AppID); err != nil {
		t.Fatal(err)
	}
	if len(official.requests) != 0 || len(store.requests) != 4 {
		t.Fatal("unexpected request routing or fallback")
	}
	for _, req := range store.requests {
		if req.path != "/api/appdetails" {
			t.Fatalf("path = %q", req.path)
		}
		assertQuery(t, req.query, "appids", "550")
	}
	for _, req := range store.requests[:3] {
		assertQuery(t, req.query, "cc", "US")
		assertQuery(t, req.query, "l", "english")
	}
}

func TestGetResolvedAppDetailsErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		appID  uint32
		status int
		body   string
		kind   sdkerrors.Kind
	}{
		{"invalid request", 0, http.StatusOK, `{}`, sdkerrors.KindRequestBuild},
		{"HTTP failure", 550, http.StatusBadGateway, `{}`, sdkerrors.KindHTTPStatus},
		{"decode failure", 550, http.StatusOK, `{`, sdkerrors.KindDecode},
		{"identity failure", 550, http.StatusOK, `{"550":{"success":false},"322070":{"success":true,"data":{"steam_appid":550}}}`, sdkerrors.KindAPIResponse},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &recordingTransport{statuses: []int{tt.status}, responseBody: tt.body}
			service := newTestService(t, &recordingTransport{}, store, 4096)
			match, err := service.GetResolvedAppDetails(context.Background(), tt.appID, nil)
			var apiErr *sdkerrors.APIError
			if !errors.As(err, &apiErr) || apiErr.Kind != tt.kind {
				t.Fatalf("error = %v, want %s", err, tt.kind)
			}
			if !reflect.DeepEqual(match, AppDetailsMatch{}) {
				t.Fatal("failure returned a match")
			}
			if tt.appID == 0 && len(store.requests) != 0 {
				t.Fatal("invalid request reached transport")
			}
		})
	}
}

func TestAppDetailsBatchDoesNotResolveIdentity(t *testing.T) {
	t.Parallel()
	store := &recordingTransport{responseBody: `{"550":{"success":false},"322070":{"success":true,"data":{"steam_appid":550}}}`}
	service := newTestService(t, &recordingTransport{}, store, 4096)
	batch, err := service.GetAppDetailsBatch(context.Background(), []uint32{550}, nil)
	if err != nil || len(batch) != 1 || batch[0].Err != nil {
		t.Fatalf("batch fetch semantics changed: %#v, %v", batch, err)
	}
	if _, err := ResolveAppDetails(batch[0].Response, batch[0].AppID); err == nil {
		t.Fatal("resolver bypassed direct failure")
	}
}
