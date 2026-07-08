package gameserversservice

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

func TestGetServerListBuildsQueryAndDecodesResponse(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		responseBody: `{
			"response": {
				"servers": [{
					"addr": "1.2.3.4:27015",
					"gameport": "27015",
					"specport": 27020,
					"steamid": 90285522207964181,
					"name": "Test Server",
					"appid": "730",
					"gamedir": "csgo",
					"version": "1.40.0",
					"product": "csgo",
					"region": "255",
					"players": "5",
					"max_players": 10,
					"bots": "1",
					"map": "de_dust2",
					"secure": "1",
					"dedicated": true,
					"os": "l",
					"gametype": "competitive"
				}]
			}
		}`,
	}
	service := newTestService(t, transport)

	filter := NewFilter().AppID(730).Secure().NotEmpty().NotFull().NoPassword()
	resp, err := service.GetServerList(context.Background(), &GetServerListOptions{
		Filter: filter.String(),
		Limit:  100,
	})
	if err != nil {
		t.Fatalf("GetServerList returned error: %v", err)
	}
	if len(resp.Response.Servers) != 1 {
		t.Fatalf("unexpected servers: %#v", resp.Response.Servers)
	}
	server := resp.Response.Servers[0]
	if server.Addr != "1.2.3.4:27015" || server.AppID != 730 || server.GamePort != 27015 || server.Players != 5 {
		t.Fatalf("unexpected server: %#v", server)
	}
	if server.SteamID != "90285522207964181" || !server.Secure || !server.Dedicated {
		t.Fatalf("unexpected server identity/security fields: %#v", server)
	}

	req := transport.onlyRequest(t)
	assertRequest(t, req, http.MethodGet, "/IGameServersService/GetServerList/v1/")
	assertQuery(t, req.query, "filter", `\appid\730\secure\1\empty\1\full\1\password\0`)
	assertQuery(t, req.query, "limit", "100")
}

func TestGetServerListAllowsNilOptions(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{responseBody: `{"response":{"servers":[]}}`}
	service := newTestService(t, transport)

	resp, err := service.GetServerList(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetServerList returned error: %v", err)
	}
	if len(resp.Response.Servers) != 0 {
		t.Fatalf("unexpected servers: %#v", resp.Response.Servers)
	}

	req := transport.onlyRequest(t)
	if got := req.query.Encode(); got != "" {
		t.Fatalf("unexpected query: %s", got)
	}
}

func TestGetServerListRawReturnsBody(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{responseBody: `{"response":{"servers":[]}}`}
	service := newTestService(t, transport)

	body, err := service.GetServerListRaw(context.Background(), &GetServerListOptions{Limit: 1})
	if err != nil {
		t.Fatalf("GetServerListRaw returned error: %v", err)
	}
	if got := strings.TrimSpace(string(body)); got != `{"response":{"servers":[]}}` {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestGetServerListValidatesFilter(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{})
	tests := []string{
		`appid\730`,
		`\appid`,
		`\appid\0`,
		`\secure\0`,
		`\password\2`,
		`\type\x`,
		`\empty\1\noplayers\1`,
		"\\appid\\730\x00",
		`\appid\730\name_match\ *surf* `,
	}
	for _, filter := range tests {
		if _, err := service.GetServerList(context.Background(), &GetServerListOptions{Filter: filter}); err == nil {
			t.Fatalf("expected validation error for %q", filter)
		}
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
		responseBody = `{"response":{"servers":[]}}`
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
