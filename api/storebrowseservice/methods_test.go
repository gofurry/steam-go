package storebrowseservice

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/gofurry/steam-go/internal/request"
)

func TestGetItemsBuildsInputJSONAndDecodesAssets(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		responseBody: `{
			"response": {
				"store_items": [{
					"item_type": 0,
					"id": 4710650,
					"success": 1,
					"visible": true,
					"name": "Fixture Game",
					"store_url_path": "app/4710650/Fixture_Game/",
					"appid": 4710650,
					"assets": {
						"asset_url_format": "steam/apps/4710650/${FILENAME}?t=1781233831",
						"library_hero_2x": "448851b668e4397d9863e571cf481b0e46e1315f/library_hero_2x.jpg"
					},
					"related_items": {"parent_appid": 1},
					"categories": [{"id": 2}]
				}]
			}
		}`,
	}
	service := newTestService(t, transport)

	resp, err := service.GetItems(context.Background(), GetItemsRequest{
		IDs: []StoreItemID{
			{AppID: 4710650},
		},
		Context: &StoreBrowseContext{
			CountryCode: " US ",
			Language:    " english ",
		},
		DataRequest: &StoreBrowseDataRequest{
			IncludeAssets: true,
		},
	})
	if err != nil {
		t.Fatalf("GetItems returned error: %v", err)
	}
	if len(resp.Response.StoreItems) != 1 {
		t.Fatalf("store item count = %d, want 1", len(resp.Response.StoreItems))
	}
	item := resp.Response.StoreItems[0]
	if item.AppID != 4710650 || item.Name != "Fixture Game" {
		t.Fatalf("unexpected item: %#v", item)
	}
	if got := item.Assets["library_hero_2x"]; !strings.HasSuffix(got, "/library_hero_2x.jpg") {
		t.Fatalf("library_hero_2x = %q", got)
	}
	if len(item.RelatedItems) == 0 || len(item.Categories) == 0 {
		t.Fatalf("raw subtrees were not retained: %#v", item)
	}

	req := transport.onlyRequest(t)
	assertRequest(t, req, http.MethodGet, "/IStoreBrowseService/GetItems/v1/")

	var input GetItemsRequest
	if err := json.Unmarshal([]byte(req.query.Get("input_json")), &input); err != nil {
		t.Fatalf("decode input_json: %v", err)
	}
	if len(input.IDs) != 1 || input.IDs[0].AppID != 4710650 {
		t.Fatalf("unexpected ids: %#v", input.IDs)
	}
	if input.Context == nil || input.Context.CountryCode != "US" || input.Context.Language != "english" {
		t.Fatalf("unexpected context: %#v", input.Context)
	}
	if input.DataRequest == nil || !input.DataRequest.IncludeAssets {
		t.Fatalf("unexpected data_request: %#v", input.DataRequest)
	}
}

func TestGetItemsValidation(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{})
	if _, err := service.GetItems(context.Background(), GetItemsRequest{}); err == nil {
		t.Fatal("expected empty ids validation error")
	}
	if _, err := service.GetItems(context.Background(), GetItemsRequest{
		IDs: []StoreItemID{{}},
	}); err == nil {
		t.Fatal("expected zero id validation error")
	}
}

func TestGetContentHubConfigRawReturnsBody(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{responseBody: `{"response":{"hubconfigs":[]}}`}
	service := newTestService(t, transport)

	body, err := service.GetContentHubConfigRaw(context.Background())
	if err != nil {
		t.Fatalf("GetContentHubConfigRaw returned error: %v", err)
	}
	if got := strings.TrimSpace(string(body)); got != `{"response":{"hubconfigs":[]}}` {
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
