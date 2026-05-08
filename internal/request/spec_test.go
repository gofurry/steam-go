package request_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/GoFurry/steam-go/internal/auth"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/request"
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
		0,
		1024,
		recorder,
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
		1,
		1024,
		recorder,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}

	_, err = executor.DoRaw(context.Background(), request.RequestSpec{
		Method:      http.MethodPost,
		Path:        "/ITestService/DoThing/v1/",
		Body:        []byte(`{"retry":true}`),
		ContentType: "application/json",
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

func TestExecutorPreservesExplicitContentTypeHeader(t *testing.T) {
	t.Parallel()

	recorder := &recordingTransport{
		statuses: []int{http.StatusOK},
	}

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		0,
		1024,
		recorder,
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
		0,
		1024,
		recorder,
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
		0,
		8,
		recorder,
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

	executor, err := request.NewExecutor("https://api.steampowered.com", nil, nil, 0, 1024, &recordingTransport{})
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

	_, err := request.NewExecutor("https://api.steampowered.com", nil, nil, 0, 0, &recordingTransport{})
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindRequestBuild {
		t.Fatalf("expected request_build error, got %v", err)
	}
}
