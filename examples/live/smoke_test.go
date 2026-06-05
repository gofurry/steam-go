package live_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/examples/live/internal/realtest"
)

const (
	liveSmokeStatusOK   = "OK"
	liveSmokeStatusWarn = "WARN"
	liveSmokeStatusFail = "FAIL"
	liveSmokeStatusSkip = "SKIP"
)

type liveSmokeReport struct {
	Summary       liveSmokeSummary `json:"summary"`
	ProxyMode     string           `json:"proxy_mode"`
	SkippedReason string           `json:"skipped_reason,omitempty"`
	Checks        []liveSmokeCheck `json:"checks"`
}

type liveSmokeSummary struct {
	OK   int `json:"ok"`
	Warn int `json:"warn"`
	Fail int `json:"fail"`
	Skip int `json:"skip"`
}

type liveSmokeCheck struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	Duration      string `json:"duration,omitempty"`
	SkippedReason string `json:"skipped_reason,omitempty"`
}

func TestLiveSmokeOptIn(t *testing.T) {
	report := newLiveSmokeReport("not_checked")

	if os.Getenv("STEAM_GO_LIVE") != "1" {
		reason := "set STEAM_GO_LIVE=1 to run live Steam smoke checks"
		report.SkippedReason = reason
		report.addCheck("live smoke opt-in", liveSmokeStatusSkip, "live smoke disabled", 0, reason)
		writeLiveSmokeReports(t, report)
		t.Skip(reason)
	}

	cfg, err := realtest.LoadConfig()
	if err != nil {
		report.addCheck("config", liveSmokeStatusFail, redactLiveSmokeMessage(err.Error()), 0, "")
		writeLiveSmokeReports(t, report)
		t.Fatalf("load live config: %v", err)
	}
	report.ProxyMode = redactedLiveSmokeProxyMode(cfg.ProxyLabel)

	client, err := realtest.NewClient(cfg)
	if err != nil {
		report.addCheck("client", liveSmokeStatusFail, redactLiveSmokeMessage(err.Error()), 0, "")
		writeLiveSmokeReports(t, report)
		t.Fatalf("create live client: %v", err)
	}
	defer client.Close()

	runLiveSmokeCheck(&report, "steamwebapiutil", func(ctx context.Context) error {
		info, err := client.API.SteamWebAPIUtil.GetServerInfo(ctx)
		if err != nil {
			return fmt.Errorf("GetServerInfo failed: %w", err)
		}
		if info.ServerTime == 0 && info.ServerTimeString == "" {
			return fmt.Errorf("empty server info response")
		}
		return nil
	})

	runLiveSmokeCheck(&report, "webstorefront", func(ctx context.Context) error {
		details, err := client.Web.Storefront.GetAppDetails(ctx, realtest.DefaultAppID, nil)
		if err != nil {
			return fmt.Errorf("GetAppDetails failed: %w", err)
		}
		app := details["550"]
		if !app.Success || app.Data.SteamAppID != realtest.DefaultAppID {
			return fmt.Errorf("unexpected appdetails response")
		}
		return nil
	})

	writeLiveSmokeReports(t, report)
	if report.Summary.Fail > 0 {
		t.Fatalf("live smoke failed: ok=%d warn=%d fail=%d skip=%d", report.Summary.OK, report.Summary.Warn, report.Summary.Fail, report.Summary.Skip)
	}
}

func newLiveSmokeReport(proxyMode string) liveSmokeReport {
	return liveSmokeReport{
		ProxyMode: proxyMode,
		Checks:    make([]liveSmokeCheck, 0, 2),
	}
}

func runLiveSmokeCheck(report *liveSmokeReport, name string, fn func(context.Context) error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := fn(ctx); err != nil {
		report.addCheck(name, liveSmokeStatusFail, redactLiveSmokeMessage(err.Error()), time.Since(start), "")
		return
	}
	report.addCheck(name, liveSmokeStatusOK, "reachable", time.Since(start), "")
}

func (r *liveSmokeReport) addCheck(name, status, message string, duration time.Duration, skippedReason string) {
	check := liveSmokeCheck{
		Name:          name,
		Status:        status,
		Message:       message,
		SkippedReason: skippedReason,
	}
	if duration > 0 {
		check.Duration = duration.Round(time.Millisecond).String()
	}
	r.Checks = append(r.Checks, check)

	switch status {
	case liveSmokeStatusOK:
		r.Summary.OK++
	case liveSmokeStatusWarn:
		r.Summary.Warn++
	case liveSmokeStatusFail:
		r.Summary.Fail++
	case liveSmokeStatusSkip:
		r.Summary.Skip++
	}
}

func writeLiveSmokeReports(t *testing.T, report liveSmokeReport) {
	t.Helper()

	if path := strings.TrimSpace(os.Getenv("STEAM_GO_LIVE_REPORT")); path != "" {
		data, err := renderLiveSmokeJSON(report)
		if err != nil {
			t.Fatalf("render live smoke JSON report: %v", err)
		}
		if err := writeLiveSmokeReportFile(path, data); err != nil {
			t.Fatalf("write live smoke JSON report: %v", err)
		}
	}

	if path := strings.TrimSpace(os.Getenv("STEAM_GO_LIVE_REPORT_HUMAN")); path != "" {
		data := []byte(renderLiveSmokeHuman(report))
		if err := writeLiveSmokeReportFile(path, data); err != nil {
			t.Fatalf("write live smoke human report: %v", err)
		}
	}
}

func writeLiveSmokeReportFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func renderLiveSmokeJSON(report liveSmokeReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

func renderLiveSmokeHuman(report liveSmokeReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "summary: ok=%d warn=%d fail=%d skip=%d proxy=%s\n", report.Summary.OK, report.Summary.Warn, report.Summary.Fail, report.Summary.Skip, report.ProxyMode)
	if report.SkippedReason != "" {
		fmt.Fprintf(&b, "skipped: %s\n", report.SkippedReason)
	}
	for _, check := range report.Checks {
		fmt.Fprintf(&b, "[%s] %s: %s", check.Status, check.Name, check.Message)
		if check.Duration != "" {
			fmt.Fprintf(&b, " duration=%s", check.Duration)
		}
		if check.SkippedReason != "" {
			fmt.Fprintf(&b, " skipped_reason=%s", check.SkippedReason)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func redactedLiveSmokeProxyMode(proxyLabel string) string {
	proxyLabel = strings.TrimSpace(proxyLabel)
	if proxyLabel == "" {
		return "direct"
	}
	return steam.RedactSensitiveURL(proxyLabel)
}

func redactLiveSmokeMessage(message string) string {
	return steam.RedactSensitiveURL(message)
}
