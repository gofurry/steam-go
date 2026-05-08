package request

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GoFurry/steam-go/internal/auth"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
)

const defaultUserAgent = "steam-go/1"

// Transport is the request executor used by the SDK.
type Transport interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// RequestSpec describes a Steam API request.
type RequestSpec struct {
	Method      string
	Path        string
	Query       url.Values
	Headers     http.Header
	Body        []byte
	ContentType string
}

// Executor performs request assembly and response reading.
type Executor struct {
	baseURL              *url.URL
	apiKeyProvider       auth.APIKeyProvider
	accessTokenProvider  auth.AccessTokenProvider
	retry                int
	maxResponseBodyBytes int64
	transport            Transport
}

type requestCredentials struct {
	apiKey      string
	accessToken string
}

// NewExecutor creates a request executor.
func NewExecutor(baseURL string, apiKeyProvider auth.APIKeyProvider, accessTokenProvider auth.AccessTokenProvider, retry int, maxResponseBodyBytes int64, transport Transport) (*Executor, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "invalid base url", nil, err)
	}
	if maxResponseBodyBytes <= 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max response body bytes must be greater than zero", nil, nil)
	}
	return &Executor{
		baseURL:              parsed,
		apiKeyProvider:       apiKeyProvider,
		accessTokenProvider:  accessTokenProvider,
		retry:                retry,
		maxResponseBodyBytes: maxResponseBodyBytes,
		transport:            transport,
	}, nil
}

// DoRaw executes the request and returns the raw response body.
func (e *Executor) DoRaw(ctx context.Context, spec RequestSpec) ([]byte, error) {
	if spec.Method == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "request method is required", nil, nil)
	}
	if spec.Path == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "request path is required", nil, nil)
	}

	relative := &url.URL{Path: spec.Path}
	resolved := e.baseURL.ResolveReference(relative)

	creds, err := e.resolveCredentials(ctx, spec.Method, resolved)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt <= e.retry; attempt++ {
		req, err := e.buildRequest(ctx, resolved, spec, creds)
		if err != nil {
			return nil, err
		}

		resp, err := e.transport.Do(ctx, req)
		if err != nil {
			lastErr = sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, err)
			if shouldRetryTransport(ctx, err) && attempt < e.retry {
				if !sleepBeforeRetry(ctx, attempt, nil) {
					return nil, sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
				}
				continue
			}
			return nil, lastErr
		}

		body, readErr := readAndCloseBody(resp, e.maxResponseBodyBytes)
		if readErr != nil {
			lastErr = sdkerrors.New(sdkerrors.KindTransport, resp.StatusCode, "read response body failed", nil, readErr)
			if attempt < e.retry {
				if !sleepBeforeRetry(ctx, attempt, resp) {
					return nil, sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
				}
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			lastErr = sdkerrors.New(
				sdkerrors.KindHTTPStatus,
				resp.StatusCode,
				fmt.Sprintf("unexpected http status %d", resp.StatusCode),
				body,
				nil,
			)
			if shouldRetryStatus(resp.StatusCode, e.apiKeyProvider != nil) && attempt < e.retry {
				if shouldRotateAPIKey(resp.StatusCode, e.apiKeyProvider != nil) {
					creds.apiKey, err = e.resolveAPIKey(req)
					if err != nil {
						return nil, err
					}
				}
				if !sleepBeforeRetry(ctx, attempt, resp) {
					return nil, sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
				}
				continue
			}
			return nil, lastErr
		}

		return body, nil
	}

	if lastErr == nil {
		lastErr = sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, nil)
	}
	return nil, lastErr
}

func (e *Executor) resolveCredentials(ctx context.Context, method string, resolved *url.URL) (requestCredentials, error) {
	req, err := http.NewRequestWithContext(ctx, method, resolved.String(), nil)
	if err != nil {
		return requestCredentials{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "build request failed", nil, err)
	}

	apiKey, err := e.resolveAPIKey(req)
	if err != nil {
		return requestCredentials{}, err
	}
	accessToken, err := e.resolveAccessToken(req)
	if err != nil {
		return requestCredentials{}, err
	}

	return requestCredentials{
		apiKey:      apiKey,
		accessToken: accessToken,
	}, nil
}

func (e *Executor) resolveAPIKey(req *http.Request) (string, error) {
	if e.apiKeyProvider == nil {
		return "", nil
	}

	apiKey, err := e.apiKeyProvider.Next(req)
	if err != nil {
		return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, "resolve api key failed", nil, err)
	}
	return apiKey, nil
}

func (e *Executor) resolveAccessToken(req *http.Request) (string, error) {
	if e.accessTokenProvider == nil {
		return "", nil
	}

	accessToken, err := e.accessTokenProvider.Next(req)
	if err != nil {
		return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, "resolve access token failed", nil, err)
	}
	return accessToken, nil
}

func (e *Executor) buildRequest(ctx context.Context, resolved *url.URL, spec RequestSpec, creds requestCredentials) (*http.Request, error) {
	var body io.Reader
	if len(spec.Body) > 0 {
		body = bytes.NewReader(spec.Body)
	}

	req, err := http.NewRequestWithContext(ctx, spec.Method, resolved.String(), body)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "build request failed", nil, err)
	}

	query := auth.InjectCredentials(spec.Query, creds.apiKey, creds.accessToken)
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", defaultUserAgent)
	for key, values := range spec.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if spec.ContentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", spec.ContentType)
	}

	return req, nil
}

func readAndCloseBody(resp *http.Response, maxBytes int64) ([]byte, error) {
	defer resp.Body.Close()
	if maxBytes <= 0 {
		return io.ReadAll(resp.Body)
	}

	reader := &io.LimitedReader{R: resp.Body, N: maxBytes + 1}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("response body exceeds limit of %d bytes", maxBytes)
	}
	return body, nil
}

func shouldRetryTransport(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

func shouldRetryStatus(statusCode int, hasAPIKeyProvider bool) bool {
	if statusCode >= http.StatusInternalServerError {
		return true
	}
	if !hasAPIKeyProvider {
		return false
	}
	return statusCode == http.StatusUnauthorized || statusCode == http.StatusTooManyRequests
}

func shouldRotateAPIKey(statusCode int, hasAPIKeyProvider bool) bool {
	if !hasAPIKeyProvider {
		return false
	}
	return statusCode == http.StatusUnauthorized || statusCode == http.StatusTooManyRequests
}

func sleepBeforeRetry(ctx context.Context, attempt int, resp *http.Response) bool {
	delay := retryDelay(attempt, resp, time.Now())
	select {
	case <-ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}

func retryDelay(attempt int, resp *http.Response, now time.Time) time.Duration {
	if delay, ok := retryAfterDelay(resp, now); ok {
		return delay
	}
	return time.Duration(attempt+1) * 100 * time.Millisecond
}

func retryAfterDelay(resp *http.Response, now time.Time) (time.Duration, bool) {
	if resp == nil {
		return 0, false
	}

	raw := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if raw == "" {
		return 0, false
	}

	seconds, err := strconv.Atoi(raw)
	if err == nil {
		if seconds < 0 {
			return 0, true
		}
		return time.Duration(seconds) * time.Second, true
	}

	when, err := http.ParseTime(raw)
	if err != nil {
		return 0, false
	}
	if when.Before(now) {
		return 0, true
	}
	return when.Sub(now), true
}
