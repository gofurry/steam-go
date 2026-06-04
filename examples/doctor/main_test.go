package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRunDoctorSuccessWithWarnSkips(t *testing.T) {
	t.Parallel()

	server := newDoctorTestServer(t, nil)
	defer server.Close()

	report := runDoctor(t.Context(), doctorConfig{
		Timeout:           time.Second,
		Retry:             0,
		BaseURL:           server.URL,
		StorefrontBaseURL: server.URL,
		CommunityBaseURL:  server.URL,
	})

	if report.Summary.Fail != 0 {
		t.Fatalf("unexpected failures: %#v", report)
	}
	if report.Summary.Warn != 4 {
		t.Fatalf("warn count = %d, want 4: %#v", report.Summary.Warn, report.Checks)
	}
	assertCheck(t, report, "official_api", "SteamWebAPIUtil.GetServerInfo", statusOK)
	assertCheck(t, report, "official_api", "SteamUser.GetPlayerSummaries", statusWarn)
	assertCheck(t, report, "web", "Community.GetInventory", statusWarn)
	if got := exitCode(report); got != 0 {
		t.Fatalf("exitCode = %d, want 0", got)
	}
}

func TestRunDoctorWithAPIKeyRunsPlayerSummaries(t *testing.T) {
	t.Parallel()

	server := newDoctorTestServer(t, nil)
	defer server.Close()

	report := runDoctor(t.Context(), doctorConfig{
		Timeout:           time.Second,
		Retry:             0,
		BaseURL:           server.URL,
		StorefrontBaseURL: server.URL,
		CommunityBaseURL:  server.URL,
		APIKey:            "secret-api-key",
		PublicInventoryID: "76561198000000001",
	})

	if report.Summary.Fail != 0 {
		t.Fatalf("unexpected failures: %#v", report)
	}
	assertCheck(t, report, "official_api", "SteamUser.GetPlayerSummaries", statusOK)
	assertCheck(t, report, "web", "Community.GetInventory", statusOK)

	var human bytes.Buffer
	renderHuman(&human, report)
	if strings.Contains(human.String(), "secret-api-key") {
		t.Fatalf("human output leaked API key:\n%s", human.String())
	}
}

func TestRunDoctorHTTPStatusFailure(t *testing.T) {
	t.Parallel()

	server := newDoctorTestServer(t, map[string]int{"/api/appdetails": http.StatusInternalServerError})
	defer server.Close()

	report := runDoctor(t.Context(), doctorConfig{
		Timeout:           time.Second,
		Retry:             0,
		BaseURL:           server.URL,
		StorefrontBaseURL: server.URL,
		CommunityBaseURL:  server.URL,
	})

	if report.Summary.Fail == 0 {
		t.Fatalf("expected failure: %#v", report)
	}
	check := assertCheck(t, report, "web", "Storefront.GetAppDetails", statusFail)
	if !strings.Contains(check.Message, "http_status status=500") {
		t.Fatalf("failure message = %q", check.Message)
	}
	if got := exitCode(report); got != 1 {
		t.Fatalf("exitCode = %d, want 1", got)
	}
}

func TestRenderJSON(t *testing.T) {
	t.Parallel()

	report := doctorReport{
		Summary: doctorSummary{OK: 1},
		Checks: []doctorCheck{{
			Category: "environment",
			Name:     "go",
			Status:   statusOK,
			Message:  "go1.test",
		}},
	}
	var out bytes.Buffer
	if err := renderJSON(&out, report); err != nil {
		t.Fatalf("renderJSON returned error: %v", err)
	}
	var decoded doctorReport
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out.String())
	}
	if decoded.Summary.OK != 1 || decoded.Checks[0].Category != "environment" {
		t.Fatalf("unexpected JSON output: %#v", decoded)
	}
}

func TestCredentialRenderingRedactsProxy(t *testing.T) {
	t.Parallel()

	report := doctorReport{}
	addCredentialChecks(&report, doctorConfig{
		APIKey:      "secret-api-key",
		AccessToken: "secret-access-token",
		ProxyURL:    "http://user:" + "proxy-secret" + "@proxy.example:8080",
	})
	finalizeSummary(&report)

	var out bytes.Buffer
	renderHuman(&out, report)
	text := out.String()
	for _, secret := range []string{"secret-api-key", "secret-access-token", "proxy-secret", "user:"} {
		if strings.Contains(text, secret) {
			t.Fatalf("credential output leaked %q:\n%s", secret, text)
		}
	}
	if !strings.Contains(text, "http://proxy.example:8080") {
		t.Fatalf("redacted proxy URL missing from output:\n%s", text)
	}
}

func newDoctorTestServer(t *testing.T, statuses map[string]int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status := statuses[r.URL.Path]; status != 0 {
			http.Error(w, "fixture error", status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/ISteamWebAPIUtil/GetServerInfo/v1/":
			_, _ = w.Write([]byte(`{"servertime":1700000000,"servertimestring":"1700000000"}`))
		case "/ISteamUser/GetPlayerSummaries/v2/":
			_, _ = w.Write([]byte(`{"response":{"players":[{"steamid":"76561198000000001","personaname":"Fixture Player"}]}}`))
		case "/api/appdetails":
			_, _ = w.Write([]byte(`{"550":{"success":true,"data":{"name":"Left 4 Dead 2 Fixture","steam_appid":550,"is_free":false,"short_description":"fixture","header_image":"header.jpg","developers":["Valve"],"publishers":["Valve"],"platforms":{"windows":true}}}}`))
		case "/appreviews/550":
			_, _ = w.Write([]byte(`{"success":1,"query_summary":{"total_reviews":12},"cursor":"fixture","reviews":[]}`))
		case "/market/priceoverview":
			_, _ = w.Write([]byte(`{"success":true,"lowest_price":"$2.30","volume":"1","median_price":"$2.40"}`))
		case "/inventory/76561198000000001/730/2":
			_, _ = w.Write([]byte(`{"success":1,"assets":[],"descriptions":[],"total_inventory_count":0,"more_items":false}`))
		default:
			http.NotFound(w, r)
		}
	}))
}

func assertCheck(t *testing.T, report doctorReport, category, name, status string) doctorCheck {
	t.Helper()
	for _, check := range report.Checks {
		if check.Category == category && check.Name == name {
			if check.Status != status {
				t.Fatalf("%s/%s status = %s, want %s: %#v", category, name, check.Status, status, check)
			}
			return check
		}
	}
	t.Fatalf("missing check %s/%s in %#v", category, name, report.Checks)
	return doctorCheck{}
}
