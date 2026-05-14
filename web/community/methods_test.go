package community

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	itraffic "github.com/gofurry/steam-go/internal/traffic"
)

func TestGetInventoryBuildsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	official := &recordingTransport{statuses: []int{http.StatusOK}}
	communityTransport := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{"success":1,"assets":[{"appid":730,"contextid":"2","assetid":"10","classid":"20","instanceid":"0","amount":"1"}],"descriptions":[{"appid":730,"classid":"20","instanceid":"0","name":"AK-47","market_name":"AK-47","market_hash_name":"AK-47 | Redline","type":"Rifle","icon_url":"icon","tradable":1,"marketable":1,"commodity":0,"actions":[{"link":"steam://"}],"descriptions":[{"type":"text","value":"field"}],"tags":[{"category":"Quality","internal_name":"Normal","localized_category_name":"Quality","localized_tag_name":"Normal"}]}],"total_inventory_count":1,"more_items":1,"last_assetid":"11"}`,
	}
	service := newTestService(t, official, communityTransport, 1024)

	resp, err := service.GetInventory(context.Background(), "76561198370695025", 730, "2", &GetInventoryOptions{
		Language:     "english",
		Count:        500,
		StartAssetID: "123",
	})
	if err != nil {
		t.Fatalf("GetInventory returned error: %v", err)
	}
	if len(resp.Assets) != 1 || len(resp.Descriptions) != 1 {
		t.Fatalf("unexpected payload sizes: %#v", resp)
	}
	if !resp.MoreItems.Bool() {
		t.Fatal("expected more_items to decode as true")
	}

	communityTransport.mu.Lock()
	defer communityTransport.mu.Unlock()
	req := communityTransport.requests[0]
	if req.path != "/inventory/76561198370695025/730/2" {
		t.Fatalf("unexpected path: %s", req.path)
	}
	assertQuery(t, req.query, "l", "english")
	assertQuery(t, req.query, "count", "500")
	assertQuery(t, req.query, "start_assetid", "123")
}

func TestGetInventoryUsesDefaultCountAndTrafficClass(t *testing.T) {
	t.Parallel()

	official := &recordingTransport{statuses: []int{http.StatusOK}}
	communityTransport := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{"success":1,"assets":[],"descriptions":[],"total_inventory_count":0,"more_items":false}`,
	}
	service := newTestService(t, official, communityTransport, 1024)

	_, err := service.GetInventory(context.Background(), "76561198370695025", 730, "2", nil)
	if err != nil {
		t.Fatalf("GetInventory returned error: %v", err)
	}

	official.mu.Lock()
	officialCount := len(official.requests)
	official.mu.Unlock()
	if officialCount != 0 {
		t.Fatalf("expected official transport to stay unused, got %d", officialCount)
	}

	communityTransport.mu.Lock()
	defer communityTransport.mu.Unlock()
	assertQuery(t, communityTransport.requests[0].query, "count", "2000")
}

func TestCommunityValidationAndErrors(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusOK}}, 16)
	if _, err := service.GetInventory(context.Background(), "", 730, "2", nil); err == nil {
		t.Fatal("expected steam id validation error")
	}
	if _, err := service.GetInventory(context.Background(), "76561198370695025", 730, "", nil); err == nil {
		t.Fatal("expected context id validation error")
	}
	if _, err := service.GetInventory(context.Background(), "76561198370695025/extra", 730, "2", nil); err == nil {
		t.Fatal("expected steam id numeric validation error")
	}
	if _, err := service.GetInventory(context.Background(), "76561198370695025", 730, "2/3", nil); err == nil {
		t.Fatal("expected context id numeric validation error")
	}

	httpErrService := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusForbidden}}, 1024)
	_, err := httpErrService.GetInventory(context.Background(), "76561198370695025", 730, "2", nil)
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindHTTPStatus {
		t.Fatalf("expected http status error, got %v", err)
	}

	decodeErrService := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{`,
	}, 1024)
	_, err = decodeErrService.GetInventory(context.Background(), "76561198370695025", 730, "2", nil)
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindDecode {
		t.Fatalf("expected decode error, got %v", err)
	}

	bodyLimitService := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: strings.Repeat("a", 32),
	}, 8)
	_, err = bodyLimitService.GetInventoryRaw(context.Background(), "76561198370695025", 730, "2", nil)
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindTransport {
		t.Fatalf("expected body limit transport error, got %v", err)
	}
}

func TestFlexibleBoolSupportsBoolAndIntPayloads(t *testing.T) {
	t.Parallel()

	var trueValue FlexibleBool
	if err := trueValue.UnmarshalJSON([]byte("1")); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if !trueValue.Bool() {
		t.Fatal("expected int payload to decode as true")
	}

	var falseValue FlexibleBool
	if err := falseValue.UnmarshalJSON([]byte("false")); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if falseValue.Bool() {
		t.Fatal("expected bool payload to decode as false")
	}
}

func newTestService(t *testing.T, officialTransport, communityTransport *recordingTransport, maxBodyBytes int64) *Service {
	t.Helper()

	executor, err := request.NewExecutor(
		"https://steamcommunity.com",
		nil,
		nil,
		maxBodyBytes,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    officialTransport,
		},
		map[itraffic.Class]request.ExecutionPolicy{
			itraffic.ClassCommunityWeb: {
				Retry:        0,
				RetryBackoff: request.DefaultRetryBackoffConfig(),
				Transport:    communityTransport,
			},
		},
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}
	return NewService(executor)
}

type recordingTransport struct {
	mu           sync.Mutex
	requests     []capturedRequest
	statuses     []int
	responseBody string
}

type capturedRequest struct {
	path  string
	query url.Values
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
	t.requests = append(t.requests, capturedRequest{path: req.URL.Path, query: clonedQuery})

	status := http.StatusOK
	if len(t.statuses) > 0 {
		status = t.statuses[0]
		t.statuses = t.statuses[1:]
	}
	body := t.responseBody
	if body == "" {
		body = "ok"
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func assertQuery(t *testing.T, query url.Values, key, want string) {
	t.Helper()
	if got := query.Get(key); got != want {
		t.Fatalf("unexpected %s: got %q want %q", key, got, want)
	}
}
