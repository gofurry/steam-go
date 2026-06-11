package contentserverdirectoryservice

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/gofurry/steam-go/internal/request"
)

func TestGetCDNForVideoBuildsQueryAndKeepsRawPayload(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		responseBody: `{"response":{"cdn":"edge","hosts":["video.test"]}}`,
	}
	service := newTestService(t, transport)

	resp, err := service.GetCDNForVideo(context.Background(), GetCDNForVideoRequest{
		PropertyType: 2,
		ClientIP:     " 203.0.113.10 ",
		ClientRegion: " US ",
	})
	if err != nil {
		t.Fatalf("GetCDNForVideo returned error: %v", err)
	}
	if !strings.Contains(string(resp.Response), `"cdn":"edge"`) {
		t.Fatalf("unexpected response: %s", resp.Response)
	}

	req := transport.onlyRequest(t)
	assertRequest(t, req, http.MethodGet, "/IContentServerDirectoryService/GetCDNForVideo/v1/")
	assertQuery(t, req.query, "property_type", "2")
	assertQuery(t, req.query, "client_ip", "203.0.113.10")
	assertQuery(t, req.query, "client_region", "US")
}

func TestGetCDNForVideoValidatesRequiredStrings(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{})
	if _, err := service.GetCDNForVideo(context.Background(), GetCDNForVideoRequest{ClientRegion: "US"}); err == nil {
		t.Fatal("expected client ip validation error")
	}
	if _, err := service.GetCDNForVideo(context.Background(), GetCDNForVideoRequest{ClientIP: "203.0.113.10"}); err == nil {
		t.Fatal("expected client region validation error")
	}
}

func TestGetClientUpdateHostsBuildsQueryAndDecodesResponse(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		responseBody: `{
			"response": {
				"hosts_kv": "\"hosts\"{}",
				"valid_until_time": 1710000000,
				"ip_country": "US"
			}
		}`,
	}
	service := newTestService(t, transport)

	resp, err := service.GetClientUpdateHosts(context.Background(), " cached ")
	if err != nil {
		t.Fatalf("GetClientUpdateHosts returned error: %v", err)
	}
	if resp.Response.HostsKV == "" || resp.Response.ValidUntilTime != 1710000000 || resp.Response.IPCountry != "US" {
		t.Fatalf("unexpected response: %#v", resp)
	}

	req := transport.onlyRequest(t)
	assertRequest(t, req, http.MethodGet, "/IContentServerDirectoryService/GetClientUpdateHosts/v1/")
	assertQuery(t, req.query, "cached_signature", "cached")
}

func TestGetDepotPatchInfoBuildsQueryAndDecodesResponse(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		responseBody: `{
			"response": {
				"is_available": true,
				"patch_size": 123,
				"patched_chunks_size": 456
			}
		}`,
	}
	service := newTestService(t, transport)

	resp, err := service.GetDepotPatchInfo(context.Background(), GetDepotPatchInfoRequest{
		AppID:            730,
		DepotID:          731,
		SourceManifestID: 1001,
		TargetManifestID: 1002,
	})
	if err != nil {
		t.Fatalf("GetDepotPatchInfo returned error: %v", err)
	}
	if !resp.Response.IsAvailable || resp.Response.PatchSize != 123 || resp.Response.PatchedChunksSize != 456 {
		t.Fatalf("unexpected response: %#v", resp)
	}

	req := transport.onlyRequest(t)
	assertRequest(t, req, http.MethodGet, "/IContentServerDirectoryService/GetDepotPatchInfo/v1/")
	assertQuery(t, req.query, "appid", "730")
	assertQuery(t, req.query, "depotid", "731")
	assertQuery(t, req.query, "source_manifestid", "1001")
	assertQuery(t, req.query, "target_manifestid", "1002")
}

func TestGetDepotPatchInfoValidation(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{})
	valid := GetDepotPatchInfoRequest{
		AppID:            730,
		DepotID:          731,
		SourceManifestID: 1001,
		TargetManifestID: 1002,
	}
	tests := []GetDepotPatchInfoRequest{
		{DepotID: valid.DepotID, SourceManifestID: valid.SourceManifestID, TargetManifestID: valid.TargetManifestID},
		{AppID: valid.AppID, SourceManifestID: valid.SourceManifestID, TargetManifestID: valid.TargetManifestID},
		{AppID: valid.AppID, DepotID: valid.DepotID, TargetManifestID: valid.TargetManifestID},
		{AppID: valid.AppID, DepotID: valid.DepotID, SourceManifestID: valid.SourceManifestID},
	}
	for _, tt := range tests {
		if _, err := service.GetDepotPatchInfo(context.Background(), tt); err == nil {
			t.Fatalf("expected validation error for %#v", tt)
		}
	}
}

func TestGetServersForSteamPipeBuildsQueryAndDecodesResponse(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		responseBody: `{
			"response": {
				"servers": [{
					"type": "CDN",
					"source_id": 1,
					"cell_id": 123,
					"load": 10,
					"weighted_load": 0.25,
					"num_entries_in_client_list": 3,
					"steam_china_only": false,
					"host": "edge.example.com",
					"vhost": "cdn.example.com",
					"use_as_proxy": true,
					"proxy_request_path_template": "/depot/{depotid}",
					"https_support": "optional",
					"allowed_app_ids": [730],
					"preferred_server": true
				}]
			}
		}`,
	}
	service := newTestService(t, transport)
	maxServers := uint32(8)
	launcherType := int32(1)

	resp, err := service.GetServersForSteamPipe(context.Background(), 123, &GetServersForSteamPipeOptions{
		MaxServers:         &maxServers,
		IPOverride:         " 203.0.113.20 ",
		LauncherType:       &launcherType,
		IPv6Public:         " 2001:db8::1 ",
		CurrentConnections: " connections-json ",
	})
	if err != nil {
		t.Fatalf("GetServersForSteamPipe returned error: %v", err)
	}
	if len(resp.Response.Servers) != 1 || resp.Response.Servers[0].Host != "edge.example.com" || !resp.Response.Servers[0].PreferredServer {
		t.Fatalf("unexpected response: %#v", resp)
	}

	req := transport.onlyRequest(t)
	assertRequest(t, req, http.MethodGet, "/IContentServerDirectoryService/GetServersForSteamPipe/v1/")
	assertQuery(t, req.query, "cell_id", "123")
	assertQuery(t, req.query, "max_servers", "8")
	assertQuery(t, req.query, "ip_override", "203.0.113.20")
	assertQuery(t, req.query, "launcher_type", "1")
	assertQuery(t, req.query, "ipv6_public", "2001:db8::1")
	assertQuery(t, req.query, "current_connections", "connections-json")
}

func TestRawMethodsReturnBody(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{responseBody: `{"response":{"ok":true}}`}
	service := newTestService(t, transport)

	body, err := service.GetClientUpdateHostsRaw(context.Background(), "")
	if err != nil {
		t.Fatalf("GetClientUpdateHostsRaw returned error: %v", err)
	}
	if got := strings.TrimSpace(string(body)); got != `{"response":{"ok":true}}` {
		t.Fatalf("unexpected body: %s", got)
	}
}

func newTestService(t *testing.T, transport *recordingTransport) *Service {
	t.Helper()

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		4096,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    transport,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}
	return NewService(executor)
}

func assertRequest(t *testing.T, req capturedRequest, method string, path string) {
	t.Helper()
	if req.method != method {
		t.Fatalf("unexpected method: %s want %s", req.method, method)
	}
	if req.path != path {
		t.Fatalf("unexpected path: %s want %s", req.path, path)
	}
}

func assertQuery(t *testing.T, query url.Values, key string, want string) {
	t.Helper()
	if got := query.Get(key); got != want {
		t.Fatalf("unexpected query %s=%q want %q", key, got, want)
	}
}

type recordingTransport struct {
	mu           sync.Mutex
	requests     []capturedRequest
	responseBody string
}

type capturedRequest struct {
	method string
	path   string
	query  url.Values
}

func (t *recordingTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	clonedQuery := make(url.Values, len(req.URL.Query()))
	for key, values := range req.URL.Query() {
		copied := make([]string, len(values))
		copy(copied, values)
		clonedQuery[key] = copied
	}
	t.requests = append(t.requests, capturedRequest{
		method: req.Method,
		path:   req.URL.Path,
		query:  clonedQuery,
	})

	responseBody := t.responseBody
	if strings.TrimSpace(responseBody) == "" {
		responseBody = `{"response":{}}`
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func (t *recordingTransport) onlyRequest(tb testing.TB) capturedRequest {
	tb.Helper()
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.requests) != 1 {
		tb.Fatalf("expected one request, got %d", len(t.requests))
	}
	return t.requests[0]
}
