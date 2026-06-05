package live_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLiveSmokeReportRendering(t *testing.T) {
	report := newLiveSmokeReport("direct")
	report.addCheck("steamwebapiutil", liveSmokeStatusOK, "reachable", 12*time.Millisecond, "")
	report.addCheck("webstorefront", liveSmokeStatusFail, "upstream returned 500", 3*time.Millisecond, "")

	data, err := renderLiveSmokeJSON(report)
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}
	var decoded liveSmokeReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if decoded.Summary.OK != 1 || decoded.Summary.Fail != 1 {
		t.Fatalf("unexpected summary: %#v", decoded.Summary)
	}

	human := renderLiveSmokeHuman(report)
	if !strings.Contains(human, "summary: ok=1 warn=0 fail=1 skip=0 proxy=direct") {
		t.Fatalf("human summary missing counts: %s", human)
	}
	if !strings.Contains(human, "[FAIL] webstorefront") {
		t.Fatalf("human report missing failed check: %s", human)
	}
}

func TestLiveSmokeSkipReportWrittenWhenOptedOut(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "live-smoke.json")
	humanPath := filepath.Join(dir, "live-smoke.txt")
	t.Setenv("STEAM_GO_LIVE_REPORT", jsonPath)
	t.Setenv("STEAM_GO_LIVE_REPORT_HUMAN", humanPath)

	reason := "set STEAM_GO_LIVE=1 to run live Steam smoke checks"
	report := newLiveSmokeReport("not_checked")
	report.SkippedReason = reason
	report.addCheck("live smoke opt-in", liveSmokeStatusSkip, "live smoke disabled", 0, reason)
	writeLiveSmokeReports(t, report)

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var decoded liveSmokeReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if decoded.Summary.Skip != 1 || decoded.SkippedReason != reason {
		t.Fatalf("unexpected skip report: %#v", decoded)
	}

	human, err := os.ReadFile(humanPath)
	if err != nil {
		t.Fatalf("read human report: %v", err)
	}
	if !strings.Contains(string(human), "[SKIP] live smoke opt-in") {
		t.Fatalf("human report missing skipped check: %s", string(human))
	}
}

func TestLiveSmokeReportRedactsProxy(t *testing.T) {
	rawProxy := "http://user:" + "proxy-secret" + "@proxy.example:8080"
	report := newLiveSmokeReport(redactedLiveSmokeProxyMode(rawProxy))
	report.addCheck("steamwebapiutil", liveSmokeStatusOK, "reachable", time.Millisecond, "")

	data, err := renderLiveSmokeJSON(report)
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}
	human := renderLiveSmokeHuman(report)
	combined := string(data) + "\n" + human

	for _, secret := range []string{"proxy-secret", "user:proxy-secret", "http://user:"} {
		if strings.Contains(combined, secret) {
			t.Fatalf("report leaked proxy credential %q: %s", secret, combined)
		}
	}
	if !strings.Contains(combined, "proxy.example:8080") {
		t.Fatalf("report lost redacted proxy host: %s", combined)
	}
}
