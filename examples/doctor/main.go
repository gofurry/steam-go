package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/web/community"
	"github.com/gofurry/steam-go/web/market"
	"github.com/gofurry/steam-go/web/storefront"
)

const (
	statusOK   = "OK"
	statusWarn = "WARN"
	statusFail = "FAIL"

	defaultTimeout        = 15 * time.Second
	defaultRetry          = 1
	defaultSteamID        = "76561198370695025"
	defaultStoreAppID     = uint32(550)
	defaultMarketAppID    = uint32(440)
	defaultMarketHashName = "Mann Co. Supply Crate Key"
	defaultInventoryAppID = uint32(730)
	defaultInventoryCtxID = "2"
)

type doctorConfig struct {
	JSON              bool
	Timeout           time.Duration
	Retry             int
	BaseURL           string
	StorefrontBaseURL string
	CommunityBaseURL  string
	APIKey            string
	AccessToken       string
	ProxyURL          string
	PublicInventoryID string
}

type doctorReport struct {
	Summary doctorSummary `json:"summary"`
	Checks  []doctorCheck `json:"checks"`
}

type doctorSummary struct {
	OK   int `json:"ok"`
	Warn int `json:"warn"`
	Fail int `json:"fail"`
}

type doctorCheck struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Detail   string `json:"detail,omitempty"`
}

func main() {
	cfg, err := configFromFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	report := runDoctor(context.Background(), cfg)
	if cfg.JSON {
		if err := renderJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "render json: %v\n", err)
			os.Exit(2)
		}
	} else {
		renderHuman(os.Stdout, report)
	}
	os.Exit(exitCode(report))
}

func configFromFlags(args []string) (doctorConfig, error) {
	cfg := doctorConfig{
		Timeout: defaultTimeout,
		Retry:   defaultRetry,
	}

	fs := flag.NewFlagSet("steam-go-doctor", flag.ContinueOnError)
	fs.BoolVar(&cfg.JSON, "json", false, "write JSON output")
	fs.DurationVar(&cfg.Timeout, "timeout", defaultTimeout, "per-check timeout")
	fs.StringVar(&cfg.BaseURL, "base-url", "", "override official Steam Web API base URL")
	fs.StringVar(&cfg.StorefrontBaseURL, "storefront-base-url", "", "override Steam Storefront base URL")
	fs.StringVar(&cfg.CommunityBaseURL, "community-base-url", "", "override Steam Community/Market base URL")
	if err := fs.Parse(args); err != nil {
		return doctorConfig{}, err
	}
	if cfg.Timeout <= 0 {
		return doctorConfig{}, fmt.Errorf("timeout must be greater than zero")
	}

	liveRoot := examplesLiveRoot()
	cfg.APIKey = envOrFile("STEAM_API_KEY", filepath.Join(liveRoot, "key.txt"))
	cfg.AccessToken = envOrFile("STEAM_ACCESS_TOKEN", filepath.Join(liveRoot, "access-token.txt"))
	cfg.ProxyURL = envOrFile("STEAM_PROXY", filepath.Join(liveRoot, "proxy.txt"))
	cfg.PublicInventoryID = strings.TrimSpace(os.Getenv("STEAM_PUBLIC_INVENTORY_ID"))
	return cfg, nil
}

func runDoctor(ctx context.Context, cfg doctorConfig) doctorReport {
	report := doctorReport{}
	addEnvironmentChecks(&report, cfg)
	addCredentialChecks(&report, cfg)

	client, ok := buildDoctorClient(&report, cfg)
	if !ok {
		finalizeSummary(&report)
		return report
	}
	defer client.Close()

	runOfficialChecks(ctx, &report, client, cfg)
	runWebChecks(ctx, &report, client, cfg)
	finalizeSummary(&report)
	return report
}

func addEnvironmentChecks(report *doctorReport, cfg doctorConfig) {
	report.add("environment", "go", statusOK, runtime.Version(), runtime.GOOS+"/"+runtime.GOARCH)
	report.add("environment", "request policy", statusOK, fmt.Sprintf("timeout=%s retry=%d", cfg.Timeout, cfg.Retry), "")
}

func addCredentialChecks(report *doctorReport, cfg doctorConfig) {
	if cfg.APIKey == "" {
		report.add("credential", "api key", statusWarn, "missing; key-backed checks will be skipped", "")
	} else {
		report.add("credential", "api key", statusOK, "present", "")
	}
	if cfg.AccessToken == "" {
		report.add("credential", "access token", statusWarn, "missing", "")
	} else {
		report.add("credential", "access token", statusOK, "present", "")
	}
	if cfg.ProxyURL == "" {
		report.add("proxy", "mode", statusOK, "direct", "")
	} else {
		report.add("proxy", "mode", statusOK, "static proxy configured", steam.RedactSensitiveURL(cfg.ProxyURL))
	}
}

func buildDoctorClient(report *doctorReport, cfg doctorConfig) (*steam.Client, bool) {
	opts := []steam.Option{
		steam.WithTimeout(cfg.Timeout),
		steam.WithRetry(cfg.Retry),
	}
	if cfg.BaseURL != "" {
		opts = append(opts, steam.WithBaseURL(cfg.BaseURL))
	}
	if cfg.StorefrontBaseURL != "" {
		opts = append(opts, steam.WithStorefrontBaseURL(cfg.StorefrontBaseURL))
	}
	if cfg.CommunityBaseURL != "" {
		opts = append(opts, steam.WithCommunityBaseURL(cfg.CommunityBaseURL))
	}
	if cfg.APIKey != "" {
		opts = append(opts, steam.WithAPIKey(cfg.APIKey))
	}
	if cfg.AccessToken != "" {
		opts = append(opts, steam.WithAccessToken(cfg.AccessToken))
	}
	if cfg.ProxyURL != "" {
		selector, err := steam.NewStaticProxySelector(cfg.ProxyURL)
		if err != nil {
			report.add("proxy", "static proxy", statusFail, "invalid proxy configuration", redactError(err))
			return nil, false
		}
		opts = append(opts, steam.WithProxySelector(selector))
	}

	client, err := steam.NewClient(opts...)
	if err != nil {
		report.add("environment", "client", statusFail, "client construction failed", redactError(err))
		return nil, false
	}
	return client, true
}

func runOfficialChecks(ctx context.Context, report *doctorReport, client *steam.Client, cfg doctorConfig) {
	withTimeout(ctx, cfg.Timeout, func(checkCtx context.Context) {
		info, err := client.API.SteamWebAPIUtil.GetServerInfo(checkCtx)
		if err != nil {
			report.add("official_api", "SteamWebAPIUtil.GetServerInfo", statusFail, classifyError(err), redactError(err))
			return
		}
		detail := info.ServerTimeString
		if detail == "" && info.ServerTime != 0 {
			detail = strconv.FormatInt(info.ServerTime, 10)
		}
		report.add("official_api", "SteamWebAPIUtil.GetServerInfo", statusOK, "official API reachable", detail)
	})

	if cfg.APIKey == "" {
		report.add("official_api", "SteamUser.GetPlayerSummaries", statusWarn, "skipped; API key missing", "")
		return
	}
	withTimeout(ctx, cfg.Timeout, func(checkCtx context.Context) {
		resp, err := client.API.SteamUser.GetPlayerSummaries(checkCtx, []string{defaultSteamID})
		if err != nil {
			report.add("official_api", "SteamUser.GetPlayerSummaries", statusFail, classifyError(err), redactError(err))
			return
		}
		report.add("official_api", "SteamUser.GetPlayerSummaries", statusOK, fmt.Sprintf("players=%d", len(resp.Response.Players)), "")
	})
}

func runWebChecks(ctx context.Context, report *doctorReport, client *steam.Client, cfg doctorConfig) {
	withTimeout(ctx, cfg.Timeout, func(checkCtx context.Context) {
		resp, err := client.Web.Storefront.GetAppDetails(checkCtx, defaultStoreAppID, &storefront.GetAppDetailsOptions{Language: "english"})
		if err != nil {
			report.add("web", "Storefront.GetAppDetails", statusFail, classifyError(err), redactError(err))
			return
		}
		app := resp[strconv.FormatUint(uint64(defaultStoreAppID), 10)]
		if !app.Success {
			report.add("web", "Storefront.GetAppDetails", statusFail, "appdetails response was not successful", "")
			return
		}
		report.add("web", "Storefront.GetAppDetails", statusOK, app.Data.Name, "")
	})

	withTimeout(ctx, cfg.Timeout, func(checkCtx context.Context) {
		resp, err := client.Web.Storefront.GetAppReviews(checkCtx, defaultStoreAppID, &storefront.GetAppReviewsOptions{NumPerPage: 1})
		if err != nil {
			report.add("web", "Storefront.GetAppReviews", statusFail, classifyError(err), redactError(err))
			return
		}
		report.add("web", "Storefront.GetAppReviews", statusOK, fmt.Sprintf("total_reviews=%d", resp.QuerySummary.TotalReviews), "")
	})

	withTimeout(ctx, cfg.Timeout, func(checkCtx context.Context) {
		resp, err := client.Web.Market.GetPriceOverview(checkCtx, defaultMarketAppID, defaultMarketHashName, &market.GetPriceOverviewOptions{Currency: 1})
		if err != nil {
			report.add("web", "Market.GetPriceOverview", statusFail, classifyError(err), redactError(err))
			return
		}
		if !resp.Success {
			report.add("web", "Market.GetPriceOverview", statusWarn, "market returned success=false", "")
			return
		}
		report.add("web", "Market.GetPriceOverview", statusOK, "market reachable", resp.LowestPrice)
	})

	if cfg.PublicInventoryID == "" {
		report.add("web", "Community.GetInventory", statusWarn, "skipped; set STEAM_PUBLIC_INVENTORY_ID for a public inventory check", "")
		return
	}
	withTimeout(ctx, cfg.Timeout, func(checkCtx context.Context) {
		resp, err := client.Web.Community.GetInventory(checkCtx, cfg.PublicInventoryID, defaultInventoryAppID, defaultInventoryCtxID, &community.GetInventoryOptions{Count: 1})
		if err != nil {
			report.add("web", "Community.GetInventory", statusFail, classifyError(err), redactError(err))
			return
		}
		report.add("web", "Community.GetInventory", statusOK, fmt.Sprintf("success=%d items=%d", resp.Success, len(resp.Assets)), "")
	})
}

func withTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context)) {
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	fn(checkCtx)
}

func classifyError(err error) string {
	var apiErr *steam.APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode > 0 {
			return fmt.Sprintf("%s status=%d", apiErr.Kind, apiErr.StatusCode)
		}
		return string(apiErr.Kind)
	}
	return "error"
}

func redactError(err error) string {
	if err == nil {
		return ""
	}
	return steam.RedactSensitiveURL(err.Error())
}

func (r *doctorReport) add(category, name, status, message, detail string) {
	r.Checks = append(r.Checks, doctorCheck{
		Category: category,
		Name:     name,
		Status:   status,
		Message:  message,
		Detail:   detail,
	})
}

func finalizeSummary(report *doctorReport) {
	for _, check := range report.Checks {
		switch check.Status {
		case statusOK:
			report.Summary.OK++
		case statusWarn:
			report.Summary.Warn++
		case statusFail:
			report.Summary.Fail++
		}
	}
}

func exitCode(report doctorReport) int {
	if report.Summary.Fail > 0 {
		return 1
	}
	return 0
}

func renderHuman(w interface{ Write([]byte) (int, error) }, report doctorReport) {
	for _, check := range report.Checks {
		line := fmt.Sprintf("[%s] %s/%s: %s", check.Status, check.Category, check.Name, check.Message)
		if check.Detail != "" {
			line += " (" + check.Detail + ")"
		}
		fmt.Fprintln(w, line)
	}
	fmt.Fprintf(w, "summary: ok=%d warn=%d fail=%d\n", report.Summary.OK, report.Summary.Warn, report.Summary.Fail)
}

func renderJSON(w interface{ Write([]byte) (int, error) }, report doctorReport) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func envOrFile(envName, path string) string {
	if value := strings.TrimSpace(os.Getenv(envName)); value != "" {
		return value
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}

func examplesLiveRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("examples", "live")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "live"))
}
