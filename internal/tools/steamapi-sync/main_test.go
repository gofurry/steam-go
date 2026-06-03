package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseOfficialEndpoints(t *testing.T) {
	data := []byte(`{
	  "apilist": {
	    "interfaces": [{
	      "name": "ISteamUser",
	      "methods": [{
	        "name": "GetPlayerSummaries",
	        "version": 2,
	        "httpmethod": "GET",
	        "parameters": [
	          {"name": "key", "type": "string", "optional": false},
	          {"name": "steamids", "type": "string", "optional": false}
	        ]
	      }]
	    }]
	  }
	}`)

	endpoints, err := parseOfficialEndpoints(data)
	if err != nil {
		t.Fatalf("parseOfficialEndpoints returned error: %v", err)
	}
	if len(endpoints) != 1 {
		t.Fatalf("endpoint count = %d, want 1", len(endpoints))
	}
	got := endpoints[0]
	if got.Key.Interface != "ISteamUser" || got.Key.Method != "GetPlayerSummaries" || got.Key.Version != 2 {
		t.Fatalf("unexpected endpoint key: %+v", got.Key)
	}
	if got.HTTPMethod != "GET" {
		t.Fatalf("HTTPMethod = %q, want GET", got.HTTPMethod)
	}
}

func TestLoadOfficialEndpointsInputDoesNotUseNetwork(t *testing.T) {
	input := filepath.Join(t.TempDir(), "supported_api.json")
	writeTestFile(t, filepath.Dir(input), filepath.Base(input), `{
	  "apilist": {
	    "interfaces": [{
	      "name": "ISteamWebAPIUtil",
	      "methods": [{"name": "GetSupportedAPIList", "version": 1, "httpmethod": "GET"}]
	    }]
	  }
	}`)

	endpoints, err := loadOfficialEndpoints(input, "http://127.0.0.1:1/unreachable", time.Millisecond, 0)
	if err != nil {
		t.Fatalf("loadOfficialEndpoints returned error: %v", err)
	}
	if len(endpoints) != 1 || endpoints[0].Key.Interface != "ISteamWebAPIUtil" {
		t.Fatalf("unexpected endpoints: %+v", endpoints)
	}
}

func TestBuildCoverageReportClassifiesEntries(t *testing.T) {
	official := []officialEndpoint{
		{Key: endpointKey{Interface: "ICovered", Method: "GetThing", Version: 1}, HTTPMethod: "GET", Parameters: []officialParameter{{Name: "key"}}},
		{Key: endpointKey{Interface: "IMissing", Method: "GetThing", Version: 1}, HTTPMethod: "GET"},
		{Key: endpointKey{Interface: "IVersioned", Method: "GetThing", Version: 2}, HTTPMethod: "POST"},
	}
	sdk := []sdkEndpoint{
		{Key: endpointKey{Interface: "ICovered", Method: "GetThing", Version: 1}, Package: "api/covered", Path: "/ICovered/GetThing/v1/", HasTyped: true, HasRaw: true},
		{Key: endpointKey{Interface: "IVersioned", Method: "GetThing", Version: 1}, Package: "api/versioned", Path: "/IVersioned/GetThing/v1/", HasTyped: true, HasRaw: true},
		{Key: endpointKey{Interface: "IExtra", Method: "GetThing", Version: 1}, Package: "api/extra", Path: "/IExtra/GetThing/v1/", HasRaw: true},
	}

	report := buildCoverageReport(official, sdk)

	wantCounts := map[string]int{
		"covered":          1,
		"missing":          1,
		"version_mismatch": 2,
		"extra_sdk":        1,
	}
	for status, want := range wantCounts {
		if got := report.StatusCounts[status]; got != want {
			t.Fatalf("status %s count = %d, want %d", status, got, want)
		}
	}
	covered := findEntry(report, "ICovered", "GetThing", 1, "covered")
	if covered.Auth != "api_key" || !covered.SDKTyped || !covered.SDKRaw {
		t.Fatalf("covered entry not annotated correctly: %+v", covered)
	}
}

func TestScanSDKEndpoints(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "go.mod", "module github.com/gofurry/steam-go\n")
	writeTestFile(t, root, filepath.Join("internal", "endpoint", "endpoint.go"), `package endpoint

const (
	SteamUserGetPlayerSummaries = "/ISteamUser/GetPlayerSummaries/v2/"
	SteamUserGetFriendList = "/ISteamUser/GetFriendList/v1/"
)
`)
	writeTestFile(t, root, filepath.Join("api", "steamuser", "service.go"), `package steamuser

import "github.com/gofurry/steam-go/internal/request"

// Service exposes ISteamUser methods.
type Service struct {
	executor *request.Executor
}
`)
	writeTestFile(t, root, filepath.Join("api", "steamuser", "methods.go"), `package steamuser

import "context"

func (s *Service) GetPlayerSummaries(ctx context.Context) (string, error) { return "", nil }
func (s *Service) GetPlayerSummariesRaw(ctx context.Context) ([]byte, error) { return nil, nil }
func (s *Service) GetFriendListRaw(ctx context.Context) ([]byte, error) { return nil, nil }
`)

	endpoints, err := scanSDKEndpoints(root)
	if err != nil {
		t.Fatalf("scanSDKEndpoints returned error: %v", err)
	}
	if len(endpoints) != 2 {
		t.Fatalf("endpoint count = %d, want 2", len(endpoints))
	}
	summaries := findSDKEndpoint(endpoints, "ISteamUser", "GetPlayerSummaries", 2)
	if summaries.Package != "api/steamuser" || !summaries.HasTyped || !summaries.HasRaw {
		t.Fatalf("GetPlayerSummaries coverage = %+v", summaries)
	}
	friends := findSDKEndpoint(endpoints, "ISteamUser", "GetFriendList", 1)
	if friends.Package != "api/steamuser" || friends.HasTyped || !friends.HasRaw {
		t.Fatalf("GetFriendList coverage = %+v", friends)
	}
}

func TestWriteReportsStable(t *testing.T) {
	report := coverageReport{
		SchemaVersion:         reportSchemaVersion,
		OfficialEndpointCount: 1,
		SDKEndpointCount:      1,
		StatusCounts:          map[string]int{"covered": 1},
		Entries: []coverageEntry{{
			Interface:  "ISteamUser",
			Method:     "GetPlayerSummaries",
			Version:    2,
			HTTPMethod: "GET",
			Auth:       "api_key",
			Status:     "covered",
			SDKPackage: "api/steamuser",
			SDKPath:    "/ISteamUser/GetPlayerSummaries/v2/",
			SDKTyped:   true,
			SDKRaw:     true,
		}},
	}
	outDir := t.TempDir()
	if err := writeReports(outDir, report); err != nil {
		t.Fatalf("writeReports returned error: %v", err)
	}

	jsonData, err := os.ReadFile(filepath.Join(outDir, "coverage.generated.json"))
	if err != nil {
		t.Fatalf("read generated JSON: %v", err)
	}
	var decoded coverageReport
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("generated JSON is invalid: %v", err)
	}
	if decoded.SchemaVersion != reportSchemaVersion {
		t.Fatalf("schema version = %d, want %d", decoded.SchemaVersion, reportSchemaVersion)
	}

	markdown, err := os.ReadFile(filepath.Join(outDir, "coverage.generated.md"))
	if err != nil {
		t.Fatalf("read generated Markdown: %v", err)
	}
	if strings.Contains(string(markdown), "Generated at") {
		t.Fatalf("generated Markdown contains a dynamic timestamp")
	}
	if !strings.Contains(string(markdown), "| ISteamUser | GetPlayerSummaries | 2 | GET | api_key | covered | api/steamuser | true | true |") {
		t.Fatalf("generated Markdown does not contain expected coverage row:\n%s", string(markdown))
	}
}

func TestResolveOutputDirAllowsAbsolutePaths(t *testing.T) {
	absolute := filepath.Join(t.TempDir(), "coverage")
	if got := resolveOutputDir("repo", absolute); got != absolute {
		t.Fatalf("absolute output dir resolved to %q, want %q", got, absolute)
	}
	if got := resolveOutputDir("repo", filepath.Join("docs", "api")); got != filepath.Join("repo", "docs", "api") {
		t.Fatalf("relative output dir resolved to %q", got)
	}
}

func findEntry(report coverageReport, iface, method string, version int, status string) coverageEntry {
	for _, entry := range report.Entries {
		if entry.Interface == iface && entry.Method == method && entry.Version == version && entry.Status == status {
			return entry
		}
	}
	return coverageEntry{}
}

func findSDKEndpoint(endpoints []sdkEndpoint, iface, method string, version int) sdkEndpoint {
	for _, endpoint := range endpoints {
		if endpoint.Key.Interface == iface && endpoint.Key.Method == method && endpoint.Key.Version == version {
			return endpoint
		}
	}
	return sdkEndpoint{}
}

func writeTestFile(t *testing.T, root, path, content string) {
	t.Helper()
	fullPath := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("create test dir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
