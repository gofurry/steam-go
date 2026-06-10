package request_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofurry/steam-go/internal/auth"
	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	"github.com/gofurry/steam-go/internal/traffic"
)

func TestExecutorSupportsRequestBody(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusOK},
	}

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	body, err := executor.DoRaw(context.Background(), request.RequestSpec{
		Method:      http.MethodPost,
		Path:        "/ITestService/DoThing/v1/",
		Query:       url.Values{"input": []string{"demo"}},
		Headers:     http.Header{"X-Test": []string{"present"}},
		Body:        []byte(`{"hello":"world"}`),
		ContentType: "application/json",
	})
	if err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf("unexpected response body: %q", string(body))
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if len(recorder.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(recorder.requests))
	}
	got := recorder.requests[0]
	if got.method != http.MethodPost {
		t.Fatalf("unexpected method: %s", got.method)
	}
	if got.body != `{"hello":"world"}` {
		t.Fatalf("unexpected body: %q", got.body)
	}
	if got.contentType != "application/json" {
		t.Fatalf("unexpected content type: %s", got.contentType)
	}
	if got.query.Get("input") != "demo" {
		t.Fatalf("unexpected query: %s", got.query.Encode())
	}
	if got.header.Get("X-Test") != "present" {
		t.Fatalf("unexpected custom header: %s", got.header.Get("X-Test"))
	}
}

func TestExecutorReusesRequestBodyAcrossRetries(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusInternalServerError, http.StatusOK},
	}

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		auth.NewStaticKeyProvider("demo-key"),
		auth.NewStaticAccessTokenProvider("demo-token"),
		1024,
		request.ExecutionPolicy{
			Retry:        1,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method:      http.MethodPost,
		Path:        "/ITestService/DoThing/v1/",
		Body:        []byte(`{"retry":true}`),
		ContentType: "application/json",
		Retryable:   request.Retryable(true),
	})
	if err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if len(recorder.requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(recorder.requests))
	}

	first := recorder.requests[0]
	second := recorder.requests[1]
	if first.body != `{"retry":true}` || second.body != `{"retry":true}` {
		t.Fatalf("expected identical bodies across retries, got %q and %q", first.body, second.body)
	}
	if first.query.Get("key") != "demo-key" || second.query.Get("key") != "demo-key" {
		t.Fatalf("expected api key to remain stable across retries")
	}
	if first.query.Get("access_token") != "demo-token" || second.query.Get("access_token") != "demo-token" {
		t.Fatalf("expected access token to remain stable across retries")
	}
}

func TestExecutorRetriesGetServerErrorByDefault(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusInternalServerError, http.StatusOK},
	}
	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        1,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	if _, err := executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/DoThing/v1/",
	}); err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if got := len(recorder.requests); got != 2 {
		t.Fatalf("expected GET retry, got %d requests", got)
	}
}

func TestExecutorDoesNotRetryPostServerErrorByDefault(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusInternalServerError, http.StatusOK},
	}
	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        1,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodPost,
		Path:   "/ITestService/DoThing/v1/",
	})
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 status error, got %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if got := len(recorder.requests); got != 1 {
		t.Fatalf("expected no POST retry by default, got %d requests", got)
	}
}

func TestExecutorRetriesPostServerErrorWhenExplicitlyRetryable(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusInternalServerError, http.StatusOK},
	}
	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        1,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	if _, err := executor.DoRaw(context.Background(), request.RequestSpec{
		Method:    http.MethodPost,
		Path:      "/ITestService/DoThing/v1/",
		Retryable: request.Retryable(true),
	}); err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if got := len(recorder.requests); got != 2 {
		t.Fatalf("expected explicit POST retry, got %d requests", got)
	}
}

func TestExecutorDoesNotRetryContextCanceledTransportError(t *testing.T) {
	t.Parallel()

	recorder := &errorTransport{err: context.Canceled}
	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        1,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/DoThing/v1/",
	})
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindTransport {
		t.Fatalf("expected transport error, got %v", err)
	}
	if got := recorder.attempts.Load(); got != 1 {
		t.Fatalf("expected no retry after context canceled, got %d attempts", got)
	}
}

func TestExecutorRotatesAccessTokenOnRateLimitRetryForRetryableRequest(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusTooManyRequests, http.StatusOK},
	}
	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		auth.NewRoundRobinAccessTokenProvider("token-a", "token-b"),
		1024,
		request.ExecutionPolicy{
			Retry:        1,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	if _, err := executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/DoThing/v1/",
	}); err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if len(recorder.requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(recorder.requests))
	}
	if got := recorder.requests[0].query.Get("access_token"); got != "token-a" {
		t.Fatalf("unexpected first access token: %s", got)
	}
	if got := recorder.requests[1].query.Get("access_token"); got != "token-b" {
		t.Fatalf("unexpected second access token: %s", got)
	}
}

func TestExecutorRotatesAccessTokenOnUnauthorizedRetry(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusUnauthorized, http.StatusOK},
	}

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		auth.NewRoundRobinAccessTokenProvider("token-a", "token-b"),
		1024,
		request.ExecutionPolicy{
			Retry:        1,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/DoThing/v1/",
	})
	if err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if len(recorder.requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(recorder.requests))
	}
	if got := recorder.requests[0].query.Get("access_token"); got != "token-a" {
		t.Fatalf("unexpected first access token: %s", got)
	}
	if got := recorder.requests[1].query.Get("access_token"); got != "token-b" {
		t.Fatalf("unexpected second access token: %s", got)
	}
}

func TestExecutorSkipsCoolingAPIKeyOnUnauthorizedRetry(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusUnauthorized, http.StatusOK},
	}

	apiKeys, err := auth.NewHealthCheckedRoundRobinKeyProvider(
		auth.KeyHealthConfig{FailureThreshold: 1, Cooldown: time.Second},
		"key-a",
		"key-b",
	)
	if err != nil {
		t.Fatalf("NewHealthCheckedRoundRobinKeyProvider returned error: %v", err)
	}

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		apiKeys,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        1,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/DoThing/v1/",
	})
	if err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if len(recorder.requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(recorder.requests))
	}
	if got := recorder.requests[0].query.Get("key"); got != "key-a" {
		t.Fatalf("unexpected first api key: %s", got)
	}
	if got := recorder.requests[1].query.Get("key"); got != "key-b" {
		t.Fatalf("unexpected second api key: %s", got)
	}
}

func TestExecutorPreservesExplicitContentTypeHeader(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusOK},
	}

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method:      http.MethodPost,
		Path:        "/ITestService/DoThing/v1/",
		Headers:     http.Header{"Content-Type": []string{"application/x-custom"}},
		Body:        []byte("payload"),
		ContentType: "application/json",
	})
	if err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if got := recorder.requests[0].contentType; got != "application/x-custom" {
		t.Fatalf("expected explicit header content type to win, got %s", got)
	}
}

func TestExecutorPreservesExplicitCredentialsFromQuery(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusOK},
	}

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		auth.NewStaticKeyProvider("global-key"),
		auth.NewStaticAccessTokenProvider("global-token"),
		1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/DoThing/v1/",
		Query: url.Values{
			"key":          []string{"explicit-key"},
			"access_token": []string{"explicit-token"},
		},
	})
	if err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if len(recorder.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(recorder.requests))
	}
	got := recorder.requests[0].query
	if got.Get("key") != "explicit-key" {
		t.Fatalf("unexpected key: %s", got.Get("key"))
	}
	if got.Get("access_token") != "explicit-token" {
		t.Fatalf("unexpected access token: %s", got.Get("access_token"))
	}
}

func TestExecutorWithoutCredentialProvidersSkipsCredentialInjection(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusOK},
	}

	executor, err := request.NewExecutor(
		"https://store.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/appreviews/550",
		Query:  url.Values{"json": []string{"1"}},
	})
	if err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if len(recorder.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(recorder.requests))
	}
	got := recorder.requests[0].query
	if got.Get("key") != "" {
		t.Fatalf("expected no api key, got %q", got.Get("key"))
	}
	if got.Get("access_token") != "" {
		t.Fatalf("expected no access token, got %q", got.Get("access_token"))
	}
}

type recordingTransport struct {
	mu           sync.Mutex
	requests     []capturedRequest
	statuses     []int
	responseBody string
}

type capturedRequest struct {
	method      string
	query       url.Values
	header      http.Header
	body        string
	contentType string
}

type errorTransport struct {
	err      error
	attempts atomic.Int32
}

func (t *errorTransport) Do(context.Context, *http.Request) (*http.Response, error) {
	t.attempts.Add(1)
	return nil, t.err
}

func (t *recordingTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_ = req.Body.Close()
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	clonedQuery := make(url.Values, len(req.URL.Query()))
	for key, values := range req.URL.Query() {
		copied := make([]string, len(values))
		copy(copied, values)
		clonedQuery[key] = copied
	}

	clonedHeader := make(http.Header, len(req.Header))
	for key, values := range req.Header {
		copied := make([]string, len(values))
		copy(copied, values)
		clonedHeader[key] = copied
	}

	t.requests = append(t.requests, capturedRequest{
		method:      req.Method,
		query:       clonedQuery,
		header:      clonedHeader,
		body:        string(body),
		contentType: req.Header.Get("Content-Type"),
	})

	status := http.StatusOK
	if len(t.statuses) > 0 {
		status = t.statuses[0]
		t.statuses = t.statuses[1:]
	}
	bodyValue := t.responseBody
	if bodyValue == "" {
		bodyValue = "ok"
	}

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(bodyValue)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestExecutorRejectsResponsesThatExceedBodyLimit(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: strings.Repeat("a", 32),
	}

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		8,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/DoThing/v1/",
	})
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindTransport {
		t.Fatalf("expected transport error for oversized body, got %v", err)
	}
}

func TestExecutorReportsRequestBuildErrorForInvalidMethod(t *testing.T) {
	t.Parallel()

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    &recordingTransport{},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{})
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindRequestBuild {
		t.Fatalf("expected request_build error, got %v", err)
	}
}

func TestExecutorRejectsInvalidBodyLimit(t *testing.T) {
	t.Parallel()

	_, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		0,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    &recordingTransport{},
		},
		nil,
	)
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindRequestBuild {
		t.Fatalf("expected request_build error, got %v", err)
	}
}

func TestExecutorRoutesTrafficClassPolicies(t *testing.T) {
	t.Parallel()

	officialTransport := &recordingTransport{
		statuses: []int{http.StatusOK},
	}
	storeTransport := &recordingTransport{
		statuses: []int{http.StatusOK},
	}
	communityTransport := &recordingTransport{
		statuses: []int{http.StatusOK},
	}
	marketTransport := &recordingTransport{
		statuses: []int{http.StatusOK},
	}

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    officialTransport,
		},
		map[traffic.Class]request.ExecutionPolicy{
			traffic.ClassPublicStorePage: {
				Retry:        0,
				RetryBackoff: request.DefaultRetryBackoffConfig(),
				Transport:    storeTransport,
			},
			traffic.ClassCommunityWeb: {
				Retry:        0,
				RetryBackoff: request.DefaultRetryBackoffConfig(),
				Transport:    communityTransport,
			},
			traffic.ClassMarketWeb: {
				Retry:        0,
				RetryBackoff: request.DefaultRetryBackoffConfig(),
				Transport:    marketTransport,
			},
		},
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	if _, err := executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/Official/v1/",
	}); err != nil {
		t.Fatalf("official request returned error: %v", err)
	}

	storeCtx := traffic.WithClass(context.Background(), traffic.ClassPublicStorePage)
	if _, err := executor.DoRaw(storeCtx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/Store/v1/",
	}); err != nil {
		t.Fatalf("store request returned error: %v", err)
	}
	communityCtx := traffic.WithClass(context.Background(), traffic.ClassCommunityWeb)
	if _, err := executor.DoRaw(communityCtx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/inventory/1/730/2",
	}); err != nil {
		t.Fatalf("community request returned error: %v", err)
	}
	marketCtx := traffic.WithClass(context.Background(), traffic.ClassMarketWeb)
	if _, err := executor.DoRaw(marketCtx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/market/priceoverview",
	}); err != nil {
		t.Fatalf("market request returned error: %v", err)
	}

	officialTransport.mu.Lock()
	officialCount := len(officialTransport.requests)
	officialTransport.mu.Unlock()
	storeTransport.mu.Lock()
	storeCount := len(storeTransport.requests)
	storeTransport.mu.Unlock()
	communityTransport.mu.Lock()
	communityCount := len(communityTransport.requests)
	communityTransport.mu.Unlock()
	marketTransport.mu.Lock()
	marketCount := len(marketTransport.requests)
	marketTransport.mu.Unlock()

	if officialCount != 1 {
		t.Fatalf("expected one official request, got %d", officialCount)
	}
	if storeCount != 1 {
		t.Fatalf("expected one store request, got %d", storeCount)
	}
	if communityCount != 1 {
		t.Fatalf("expected one community request, got %d", communityCount)
	}
	if marketCount != 1 {
		t.Fatalf("expected one market request, got %d", marketCount)
	}
}

func TestExecutorAppliesPrepareRequestHook(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusOK},
	}

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			PrepareRequest: func(req *http.Request) error {
				req.Header.Set("X-Prepared", "yes")
				return nil
			},
			Transport: recorder,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	if _, err := executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/Prepared/v1/",
	}); err != nil {
		t.Fatalf("DoRaw returned error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if got := recorder.requests[0].header.Get("X-Prepared"); got != "yes" {
		t.Fatalf("expected prepared header, got %q", got)
	}
}

func TestExecutorReportsPrepareRequestErrorAsRequestBuild(t *testing.T) {
	t.Parallel()

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			PrepareRequest: func(*http.Request) error {
				return errors.New("prepare failed")
			},
			Transport: &recordingTransport{},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method: http.MethodGet,
		Path:   "/ITestService/Prepared/v1/",
	})
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindRequestBuild {
		t.Fatalf("expected request_build error, got %v", err)
	}
}
