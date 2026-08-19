package steamuser

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

func TestResolveVanityURLBuildsDefaultQueryAndDecodes(t *testing.T) {
	t.Parallel()
	transport := &recordingTransport{responseBody: `{"response":{"steamid":"76561198000000001","success":1}}`}
	service := newTestService(t, transport)

	response, err := service.ResolveVanityURL(context.Background(), "  example-user  ", nil)
	if err != nil {
		t.Fatalf("ResolveVanityURL returned error: %v", err)
	}
	if response.Response.SteamID != "76561198000000001" || response.Response.Success != 1 {
		t.Fatalf("unexpected response: %#v", response)
	}

	request := transport.onlyRequest(t)
	if request.method != http.MethodGet || request.path != "/ISteamUser/ResolveVanityURL/v1/" {
		t.Fatalf("unexpected request: %#v", request)
	}
	if got := request.query.Get("vanityurl"); got != "example-user" {
		t.Fatalf("vanityurl = %q", got)
	}
	if _, ok := request.query["url_type"]; ok {
		t.Fatalf("default request unexpectedly sent url_type: %s", request.query.Encode())
	}
}

func TestResolveVanityURLTypes(t *testing.T) {
	t.Parallel()
	for _, urlType := range []VanityURLType{VanityURLTypeIndividual, VanityURLTypeGroup, VanityURLTypeOfficialGameGroup} {
		urlType := urlType
		t.Run(string(rune('0'+urlType)), func(t *testing.T) {
			t.Parallel()
			transport := &recordingTransport{responseBody: `{"response":{"success":42,"message":"No match"}}`}
			service := newTestService(t, transport)
			response, err := service.ResolveVanityURL(context.Background(), "missing", &ResolveVanityURLOptions{URLType: urlType})
			if err != nil {
				t.Fatalf("ResolveVanityURL returned error: %v", err)
			}
			if response.Response.Success != 42 || response.Response.Message != "No match" || response.Response.SteamID != "" {
				t.Fatalf("unexpected not-found response: %#v", response)
			}
			if got := transport.onlyRequest(t).query.Get("url_type"); got != string(rune('0'+urlType)) {
				t.Fatalf("url_type = %q", got)
			}
		})
	}
}

func TestResolveVanityURLRaw(t *testing.T) {
	t.Parallel()
	transport := &recordingTransport{responseBody: `{"response":{"success":42}}`}
	body, err := newTestService(t, transport).ResolveVanityURLRaw(context.Background(), "missing", &ResolveVanityURLOptions{})
	if err != nil {
		t.Fatalf("ResolveVanityURLRaw returned error: %v", err)
	}
	if got := strings.TrimSpace(string(body)); got != `{"response":{"success":42}}` {
		t.Fatalf("unexpected raw body: %s", got)
	}
	if _, ok := transport.onlyRequest(t).query["url_type"]; ok {
		t.Fatal("zero url type should be omitted")
	}
}

func TestResolveVanityURLValidation(t *testing.T) {
	t.Parallel()
	service := newTestService(t, &recordingTransport{})
	if _, err := service.ResolveVanityURL(context.Background(), "  ", nil); err == nil {
		t.Fatal("expected empty vanity validation error")
	}
	for _, urlType := range []VanityURLType{-1, 4, 100} {
		if _, err := service.ResolveVanityURL(context.Background(), "example", &ResolveVanityURLOptions{URLType: urlType}); err == nil {
			t.Fatalf("expected validation error for type %d", urlType)
		}
	}
}

func TestVanityURLTypeConstants(t *testing.T) {
	t.Parallel()
	if VanityURLTypeIndividual != 1 || VanityURLTypeGroup != 2 || VanityURLTypeOfficialGameGroup != 3 {
		t.Fatalf("unexpected vanity URL type constants")
	}
}

func newTestService(t *testing.T, transport *recordingTransport) *Service {
	t.Helper()
	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		4096,
		request.ExecutionPolicy{Retry: 0, RetryBackoff: request.DefaultRetryBackoffConfig(), Transport: transport},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}
	return NewService(executor)
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

func (transport *recordingTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	transport.mu.Lock()
	defer transport.mu.Unlock()
	query := make(url.Values, len(req.URL.Query()))
	for key, values := range req.URL.Query() {
		query[key] = append([]string(nil), values...)
	}
	transport.requests = append(transport.requests, capturedRequest{method: req.Method, path: req.URL.Path, query: query})
	body := transport.responseBody
	if body == "" {
		body = `{"response":{}}`
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func (transport *recordingTransport) onlyRequest(t *testing.T) capturedRequest {
	t.Helper()
	transport.mu.Lock()
	defer transport.mu.Unlock()
	if len(transport.requests) != 1 {
		t.Fatalf("request count = %d, want 1", len(transport.requests))
	}
	return transport.requests[0]
}
