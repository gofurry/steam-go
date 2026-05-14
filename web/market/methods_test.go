package market

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

func TestGetPriceOverviewBuildsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	official := &recordingTransport{statuses: []int{http.StatusOK}}
	marketTransport := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{"success":true,"lowest_price":"$2.30","volume":"1,234","median_price":"$2.40"}`,
	}
	service := newTestService(t, official, marketTransport, 1024)

	resp, err := service.GetPriceOverview(context.Background(), 440, "Mann Co. Supply Crate Key", nil)
	if err != nil {
		t.Fatalf("GetPriceOverview returned error: %v", err)
	}
	if got := resp.LowestPrice; got != "$2.30" {
		t.Fatalf("unexpected lowest price: %q", got)
	}

	official.mu.Lock()
	officialCount := len(official.requests)
	official.mu.Unlock()
	if officialCount != 0 {
		t.Fatalf("expected official transport to stay unused, got %d", officialCount)
	}

	marketTransport.mu.Lock()
	defer marketTransport.mu.Unlock()
	req := marketTransport.requests[0]
	if req.path != "/market/priceoverview" {
		t.Fatalf("unexpected path: %s", req.path)
	}
	assertQuery(t, req.query, "appid", "440")
	assertQuery(t, req.query, "market_hash_name", "Mann Co. Supply Crate Key")
	assertQuery(t, req.query, "currency", "1")
}

func TestGetPriceOverviewValidationAndErrors(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusOK}}, 16)
	if _, err := service.GetPriceOverview(context.Background(), 0, "Key", nil); err == nil {
		t.Fatal("expected app id validation error")
	}
	if _, err := service.GetPriceOverview(context.Background(), 440, "", nil); err == nil {
		t.Fatal("expected market hash name validation error")
	}

	httpErrService := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusTooManyRequests}}, 1024)
	_, err := httpErrService.GetPriceOverview(context.Background(), 440, "Key", nil)
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindHTTPStatus {
		t.Fatalf("expected http status error, got %v", err)
	}

	decodeErrService := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{`,
	}, 1024)
	_, err = decodeErrService.GetPriceOverview(context.Background(), 440, "Key", nil)
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindDecode {
		t.Fatalf("expected decode error, got %v", err)
	}

	bodyLimitService := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: strings.Repeat("a", 32),
	}, 8)
	_, err = bodyLimitService.GetPriceOverviewRaw(context.Background(), 440, "Key", nil)
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindTransport {
		t.Fatalf("expected body limit transport error, got %v", err)
	}
}

func newTestService(t *testing.T, officialTransport, marketTransport *recordingTransport, maxBodyBytes int64) *Service {
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
			itraffic.ClassMarketWeb: {
				Retry:        0,
				RetryBackoff: request.DefaultRetryBackoffConfig(),
				Transport:    marketTransport,
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
