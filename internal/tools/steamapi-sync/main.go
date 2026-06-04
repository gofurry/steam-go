package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultOfficialAPIListURL = "https://api.steampowered.com/ISteamWebAPIUtil/GetSupportedAPIList/v1/?format=json"
	defaultOutputDir          = "docs/api"
	reportSchemaVersion       = 1
)

var endpointPathPattern = regexp.MustCompile(`^/([^/]+)/([^/]+)/v([0-9]+)/$`)

func main() {
	var (
		inputPath string
		outputDir string
		sourceURL string
		timeout   time.Duration
		retries   int
	)
	flag.StringVar(&inputPath, "input", "", "read GetSupportedAPIList JSON from a local file instead of the network")
	flag.StringVar(&outputDir, "output-dir", defaultOutputDir, "directory for coverage.generated.md, coverage.generated.json, and coverage-diff.md")
	flag.StringVar(&sourceURL, "url", defaultOfficialAPIListURL, "GetSupportedAPIList JSON URL")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "network timeout when -input is not set")
	flag.IntVar(&retries, "retries", 2, "network retries when -input is not set")
	flag.Parse()

	root, err := findRepoRoot()
	if err != nil {
		fatalf("%v", err)
	}

	official, err := loadOfficialEndpoints(inputPath, sourceURL, timeout, retries)
	if err != nil {
		fatalf("load official API inventory: %v", err)
	}
	sdk, err := scanSDKEndpoints(root)
	if err != nil {
		fatalf("scan SDK coverage: %v", err)
	}

	report := buildCoverageReport(official, sdk)
	if err := writeReports(resolveOutputDir(root, outputDir), report); err != nil {
		fatalf("write reports: %v", err)
	}

	fmt.Printf("Wrote Steam API coverage reports to %s\n", outputDir)
	fmt.Printf("Official endpoints: %d, SDK endpoints: %d\n", report.OfficialEndpointCount, report.SDKEndpointCount)
}

type apiListPayload struct {
	APIList struct {
		Interfaces []officialInterface `json:"interfaces"`
	} `json:"apilist"`
}

type officialInterface struct {
	Name    string           `json:"name"`
	Methods []officialMethod `json:"methods"`
}

type officialMethod struct {
	Name        string              `json:"name"`
	Version     int                 `json:"version"`
	HTTPMethod  string              `json:"httpmethod"`
	Description string              `json:"description"`
	Parameters  []officialParameter `json:"parameters"`
}

type officialParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
	Description string `json:"description"`
}

type endpointKey struct {
	Interface string
	Method    string
	Version   int
}

func (k endpointKey) pair() string {
	return k.Interface + "/" + k.Method
}

type officialEndpoint struct {
	Key        endpointKey
	HTTPMethod string
	Parameters []officialParameter
}

type sdkEndpoint struct {
	Key       endpointKey
	ConstName string
	Path      string
	Package   string
	HasTyped  bool
	HasRaw    bool
}

type methodCoverage struct {
	HasTyped bool
	HasRaw   bool
}

type serviceCoverage struct {
	Package string
	Methods map[string]methodCoverage
}

type coverageReport struct {
	SchemaVersion         int             `json:"schema_version"`
	OfficialEndpointCount int             `json:"official_endpoint_count"`
	SDKEndpointCount      int             `json:"sdk_endpoint_count"`
	StatusCounts          map[string]int  `json:"status_counts"`
	Entries               []coverageEntry `json:"entries"`
}

type coverageEntry struct {
	Interface  string              `json:"interface"`
	Method     string              `json:"method"`
	Version    int                 `json:"version"`
	HTTPMethod string              `json:"http_method,omitempty"`
	Auth       string              `json:"auth"`
	Status     string              `json:"status"`
	Parameters []coverageParameter `json:"parameters,omitempty"`
	SDKPackage string              `json:"sdk_package,omitempty"`
	SDKPath    string              `json:"sdk_path,omitempty"`
	SDKTyped   bool                `json:"sdk_typed"`
	SDKRaw     bool                `json:"sdk_raw"`
}

type coverageParameter struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Optional bool   `json:"optional"`
}

func loadOfficialEndpoints(inputPath, sourceURL string, timeout time.Duration, retries int) ([]officialEndpoint, error) {
	var data []byte
	var err error
	if strings.TrimSpace(inputPath) != "" {
		data, err = os.ReadFile(inputPath)
	} else {
		data, err = fetchURL(sourceURL, timeout, retries)
	}
	if err != nil {
		return nil, err
	}
	return parseOfficialEndpoints(data)
}

func fetchURL(sourceURL string, timeout time.Duration, retries int) ([]byte, error) {
	attempts := retries + 1
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		data, err := fetchURLOnce(sourceURL, timeout)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if attempt < attempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	return nil, lastErr
}

func fetchURLOnce(sourceURL string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(sourceURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("GET %s returned %s", sourceURL, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func parseOfficialEndpoints(data []byte) ([]officialEndpoint, error) {
	var payload apiListPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	var endpoints []officialEndpoint
	for _, iface := range payload.APIList.Interfaces {
		interfaceName := strings.TrimSpace(iface.Name)
		if interfaceName == "" {
			continue
		}
		for _, method := range iface.Methods {
			methodName := strings.TrimSpace(method.Name)
			if methodName == "" || method.Version <= 0 {
				continue
			}
			endpoints = append(endpoints, officialEndpoint{
				Key: endpointKey{
					Interface: interfaceName,
					Method:    methodName,
					Version:   method.Version,
				},
				HTTPMethod: strings.ToUpper(strings.TrimSpace(method.HTTPMethod)),
				Parameters: method.Parameters,
			})
		}
	}
	sortOfficialEndpoints(endpoints)
	return endpoints, nil
}

func scanSDKEndpoints(root string) ([]sdkEndpoint, error) {
	endpoints, err := parseEndpointConstants(filepath.Join(root, "internal", "endpoint", "endpoint.go"))
	if err != nil {
		return nil, err
	}
	services, err := parseServiceCoverage(filepath.Join(root, "api"))
	if err != nil {
		return nil, err
	}
	for i := range endpoints {
		service := services[endpoints[i].Key.Interface]
		endpoints[i].Package = service.Package
		method := service.Methods[endpoints[i].Key.Method]
		endpoints[i].HasTyped = method.HasTyped
		endpoints[i].HasRaw = method.HasRaw
	}
	sortSDKEndpoints(endpoints)
	return endpoints, nil
}

func parseEndpointConstants(path string) ([]sdkEndpoint, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}
	var endpoints []sdkEndpoint
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range valueSpec.Names {
				if i >= len(valueSpec.Values) {
					continue
				}
				lit, ok := valueSpec.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				pathValue, err := strconv.Unquote(lit.Value)
				if err != nil {
					return nil, err
				}
				key, ok := parseEndpointPath(pathValue)
				if !ok {
					continue
				}
				endpoints = append(endpoints, sdkEndpoint{
					Key:       key,
					ConstName: name.Name,
					Path:      pathValue,
				})
			}
		}
	}
	return endpoints, nil
}

func parseEndpointPath(pathValue string) (endpointKey, bool) {
	matches := endpointPathPattern.FindStringSubmatch(pathValue)
	if matches == nil {
		return endpointKey{}, false
	}
	version, err := strconv.Atoi(matches[3])
	if err != nil {
		return endpointKey{}, false
	}
	return endpointKey{Interface: matches[1], Method: matches[2], Version: version}, true
}

func parseServiceCoverage(apiDir string) (map[string]serviceCoverage, error) {
	entries, err := os.ReadDir(apiDir)
	if err != nil {
		return nil, err
	}
	services := map[string]serviceCoverage{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(apiDir, entry.Name())
		interfaceName, err := parseServiceInterface(filepath.Join(dir, "service.go"))
		if err != nil {
			return nil, err
		}
		methods, err := parseServiceMethods(filepath.Join(dir, "methods.go"))
		if err != nil {
			return nil, err
		}
		services[interfaceName] = serviceCoverage{
			Package: "api/" + entry.Name(),
			Methods: methods,
		}
	}
	return services, nil
}

var serviceInterfacePattern = regexp.MustCompile(`Service exposes ([A-Za-z0-9_]+) methods\.`)

func parseServiceInterface(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	matches := serviceInterfacePattern.FindSubmatch(data)
	if matches == nil {
		return "", fmt.Errorf("cannot find service interface comment in %s", path)
	}
	return string(matches[1]), nil
}

func parseServiceMethods(path string) (map[string]methodCoverage, error) {
	methods := map[string]methodCoverage{}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || !fn.Name.IsExported() || !hasServiceReceiver(fn) {
			continue
		}
		methodName := fn.Name.Name
		key := strings.TrimSuffix(methodName, "Raw")
		coverage := methods[key]
		if strings.HasSuffix(methodName, "Raw") {
			coverage.HasRaw = true
		} else {
			coverage.HasTyped = true
		}
		methods[key] = coverage
	}
	return methods, nil
}

func hasServiceReceiver(fn *ast.FuncDecl) bool {
	if len(fn.Recv.List) == 0 {
		return false
	}
	typ := fn.Recv.List[0].Type
	if star, ok := typ.(*ast.StarExpr); ok {
		typ = star.X
	}
	ident, ok := typ.(*ast.Ident)
	return ok && ident.Name == "Service"
}

func buildCoverageReport(official []officialEndpoint, sdk []sdkEndpoint) coverageReport {
	officialExact := map[endpointKey]officialEndpoint{}
	officialPairs := map[string][]officialEndpoint{}
	for _, endpoint := range official {
		officialExact[endpoint.Key] = endpoint
		officialPairs[endpoint.Key.pair()] = append(officialPairs[endpoint.Key.pair()], endpoint)
	}

	sdkExact := map[endpointKey]sdkEndpoint{}
	sdkPairs := map[string][]sdkEndpoint{}
	for _, endpoint := range sdk {
		sdkExact[endpoint.Key] = endpoint
		sdkPairs[endpoint.Key.pair()] = append(sdkPairs[endpoint.Key.pair()], endpoint)
	}

	var entries []coverageEntry
	for _, endpoint := range official {
		entry := newCoverageEntry(endpoint)
		if sdkEndpoint, ok := sdkExact[endpoint.Key]; ok && sdkEndpoint.isExposed() {
			entry.Status = "covered"
			applySDKCoverage(&entry, sdkEndpoint)
		} else if sdkEndpoint, ok := sdkExact[endpoint.Key]; ok {
			entry.Status = "missing"
			applySDKCoverage(&entry, sdkEndpoint)
		} else if candidates := sdkPairs[endpoint.Key.pair()]; len(candidates) > 0 {
			entry.Status = "version_mismatch"
			applySDKCoverage(&entry, candidates[0])
		} else {
			entry.Status = "missing"
		}
		entries = append(entries, entry)
	}

	for _, endpoint := range sdk {
		if _, ok := officialExact[endpoint.Key]; ok {
			continue
		}
		status := "extra_sdk"
		if len(officialPairs[endpoint.Key.pair()]) > 0 {
			status = "version_mismatch"
		}
		entry := coverageEntry{
			Interface: endpoint.Key.Interface,
			Method:    endpoint.Key.Method,
			Version:   endpoint.Key.Version,
			Auth:      "unknown",
			Status:    status,
		}
		applySDKCoverage(&entry, endpoint)
		entries = append(entries, entry)
	}

	sortCoverageEntries(entries)
	statusCounts := map[string]int{}
	for _, entry := range entries {
		statusCounts[entry.Status]++
	}

	return coverageReport{
		SchemaVersion:         reportSchemaVersion,
		OfficialEndpointCount: len(official),
		SDKEndpointCount:      len(sdk),
		StatusCounts:          statusCounts,
		Entries:               entries,
	}
}

func newCoverageEntry(endpoint officialEndpoint) coverageEntry {
	params := make([]coverageParameter, 0, len(endpoint.Parameters))
	for _, param := range endpoint.Parameters {
		params = append(params, coverageParameter{
			Name:     strings.TrimSpace(param.Name),
			Type:     strings.TrimSpace(param.Type),
			Optional: param.Optional,
		})
	}
	return coverageEntry{
		Interface:  endpoint.Key.Interface,
		Method:     endpoint.Key.Method,
		Version:    endpoint.Key.Version,
		HTTPMethod: endpoint.HTTPMethod,
		Auth:       inferAuth(endpoint.Parameters),
		Parameters: params,
	}
}

func applySDKCoverage(entry *coverageEntry, endpoint sdkEndpoint) {
	entry.SDKPackage = endpoint.Package
	entry.SDKPath = endpoint.Path
	entry.SDKTyped = endpoint.HasTyped
	entry.SDKRaw = endpoint.HasRaw
}

func (e sdkEndpoint) isExposed() bool {
	return e.HasTyped || e.HasRaw
}

func inferAuth(params []officialParameter) string {
	var kinds []string
	seen := map[string]bool{}
	for _, param := range params {
		name := strings.ToLower(strings.TrimSpace(param.Name))
		var kind string
		switch name {
		case "key":
			kind = "api_key"
		case "access_token":
			kind = "access_token"
		case "webapi_token":
			kind = "webapi_token"
		case "sessionid", "steamloginsecure":
			kind = "session"
		default:
			if strings.Contains(name, "token") {
				kind = "token"
			}
		}
		if kind != "" && !seen[kind] {
			kinds = append(kinds, kind)
			seen[kind] = true
		}
	}
	if len(kinds) == 0 {
		return "unknown"
	}
	sort.Strings(kinds)
	return strings.Join(kinds, ",")
}

func writeReports(outputDir string, report coverageReport) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	jsonData, err := marshalStableJSON(report)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "coverage.generated.json"), jsonData, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "coverage.generated.md"), []byte(renderCoverageMarkdown(report)), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "coverage-diff.md"), []byte(renderCoverageDiffMarkdown(report)), 0o644); err != nil {
		return err
	}
	return nil
}

func resolveOutputDir(root, outputDir string) string {
	if filepath.IsAbs(outputDir) {
		return outputDir
	}
	return filepath.Join(root, outputDir)
}

func marshalStableJSON(report coverageReport) ([]byte, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func renderCoverageMarkdown(report coverageReport) string {
	var b strings.Builder
	writeReportHeader(&b, "Steam Web API Coverage")
	fmt.Fprintf(&b, "- Official endpoints: %d\n", report.OfficialEndpointCount)
	fmt.Fprintf(&b, "- SDK endpoints: %d\n", report.SDKEndpointCount)
	writeStatusCounts(&b, report)
	b.WriteString("\n| Interface | Method | Version | HTTP | Auth | Status | SDK | Typed | Raw |\n")
	b.WriteString("|---|---|---:|---|---|---|---|---|---|\n")
	for _, entry := range report.Entries {
		fmt.Fprintf(
			&b,
			"| %s | %s | %d | %s | %s | %s | %s | %t | %t |\n",
			md(entry.Interface),
			md(entry.Method),
			entry.Version,
			md(entry.HTTPMethod),
			md(entry.Auth),
			md(entry.Status),
			md(entry.SDKPackage),
			entry.SDKTyped,
			entry.SDKRaw,
		)
	}
	return b.String()
}

func renderCoverageDiffMarkdown(report coverageReport) string {
	var b strings.Builder
	writeReportHeader(&b, "Steam Web API Coverage Diff")
	writeStatusCounts(&b, report)
	b.WriteString("\n| Interface | Method | Version | Status | SDK | SDK Path |\n")
	b.WriteString("|---|---|---:|---|---|---|\n")
	count := 0
	for _, entry := range report.Entries {
		if entry.Status == "covered" {
			continue
		}
		count++
		fmt.Fprintf(
			&b,
			"| %s | %s | %d | %s | %s | %s |\n",
			md(entry.Interface),
			md(entry.Method),
			entry.Version,
			md(entry.Status),
			md(entry.SDKPackage),
			md(entry.SDKPath),
		)
	}
	if count == 0 {
		b.WriteString("\nNo coverage differences found.\n")
	}
	return b.String()
}

func writeReportHeader(b *strings.Builder, title string) {
	fmt.Fprintf(b, "# %s\n\n", title)
	b.WriteString("Generated by `go run ./internal/tools/steamapi-sync -output-dir docs/api`.\n\n")
	b.WriteString("Do not edit this file manually.\n\n")
}

func writeStatusCounts(b *strings.Builder, report coverageReport) {
	statuses := make([]string, 0, len(report.StatusCounts))
	for status := range report.StatusCounts {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)
	if len(statuses) == 0 {
		return
	}
	b.WriteString("- Status counts:")
	for _, status := range statuses {
		fmt.Fprintf(b, " `%s=%d`", status, report.StatusCounts[status])
	}
	b.WriteString("\n")
}

func md(value string) string {
	if value == "" {
		return "-"
	}
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.ReplaceAll(value, "|", `\|`)
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		goMod := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(goMod)
		if err == nil && bytes.Contains(data, []byte("module github.com/gofurry/steam-go")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find steam-go repository root")
		}
		dir = parent
	}
}

func sortOfficialEndpoints(endpoints []officialEndpoint) {
	sort.Slice(endpoints, func(i, j int) bool {
		return compareKeys(endpoints[i].Key, endpoints[j].Key) < 0
	})
}

func sortSDKEndpoints(endpoints []sdkEndpoint) {
	sort.Slice(endpoints, func(i, j int) bool {
		return compareKeys(endpoints[i].Key, endpoints[j].Key) < 0
	})
}

func sortCoverageEntries(entries []coverageEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return compareKeys(
			endpointKey{Interface: entries[i].Interface, Method: entries[i].Method, Version: entries[i].Version},
			endpointKey{Interface: entries[j].Interface, Method: entries[j].Method, Version: entries[j].Version},
		) < 0
	})
}

func compareKeys(a, b endpointKey) int {
	if a.Interface != b.Interface {
		return strings.Compare(a.Interface, b.Interface)
	}
	if a.Method != b.Method {
		return strings.Compare(a.Method, b.Method)
	}
	if a.Version < b.Version {
		return -1
	}
	if a.Version > b.Version {
		return 1
	}
	return 0
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
