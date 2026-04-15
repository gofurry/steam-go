package request

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/GoFurry/steam-go/internal/auth"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
)

const defaultUserAgent = "steam-go/1"

// Logger matches the root logger contract.
type Logger interface {
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
}

// Transport is the request executor used by the SDK.
type Transport interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// RequestSpec describes a Steam API request.
type RequestSpec struct {
	Method  string
	Path    string
	Query   url.Values
	Headers http.Header
}

// Executor performs request assembly and response reading.
type Executor struct {
	baseURL             *url.URL
	apiKeyProvider      auth.APIKeyProvider
	accessTokenProvider auth.AccessTokenProvider
	retry               int
	transport           Transport
	logger              Logger
}

// NewExecutor creates a request executor.
func NewExecutor(baseURL string, apiKeyProvider auth.APIKeyProvider, accessTokenProvider auth.AccessTokenProvider, retry int, transport Transport, logger Logger) (*Executor, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "invalid base url", nil, err)
	}
	return &Executor{
		baseURL:             parsed,
		apiKeyProvider:      apiKeyProvider,
		accessTokenProvider: accessTokenProvider,
		retry:               retry,
		transport:           transport,
		logger:              logger,
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

	var lastErr error
	for attempt := 0; attempt <= e.retry; attempt++ {
		req, err := e.buildRequest(ctx, resolved, spec)
		if err != nil {
			return nil, err
		}

		resp, err := e.transport.Do(ctx, req)
		if err != nil {
			lastErr = sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, err)
			if shouldRetryTransport(ctx, err) && attempt < e.retry {
				if !sleepBeforeRetry(ctx, attempt) {
					return nil, sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
				}
				continue
			}
			return nil, lastErr
		}

		body, readErr := readAndCloseBody(resp)
		if readErr != nil {
			lastErr = sdkerrors.New(sdkerrors.KindTransport, resp.StatusCode, "read response body failed", nil, readErr)
			if attempt < e.retry {
				if !sleepBeforeRetry(ctx, attempt) {
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
				if !sleepBeforeRetry(ctx, attempt) {
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

func (e *Executor) buildRequest(ctx context.Context, resolved *url.URL, spec RequestSpec) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, spec.Method, resolved.String(), nil)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "build request failed", nil, err)
	}

	apiKey, err := e.resolveAPIKey(req)
	if err != nil {
		return nil, err
	}
	accessToken, err := e.resolveAccessToken(req)
	if err != nil {
		return nil, err
	}
	query := auth.InjectCredentials(spec.Query, apiKey, accessToken)
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", defaultUserAgent)
	for key, values := range spec.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return req, nil
}

func readAndCloseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
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

func sleepBeforeRetry(ctx context.Context, attempt int) bool {
	delay := time.Duration(attempt+1) * 100 * time.Millisecond
	select {
	case <-ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}
