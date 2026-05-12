package request

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GoFurry/steam-go/internal/auth"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/traffic"
)

const defaultUserAgent = "steam-go/1"

// Transport is the request executor used by the SDK.
type Transport interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// RequestSpec describes a Steam API request.
type RequestSpec struct {
	Method       string
	Path         string
	Query        url.Values
	Headers      http.Header
	Body         []byte
	ContentType  string
	TrafficClass traffic.Class
}

// Executor performs request assembly and response reading.
type Executor struct {
	baseURL              *url.URL
	apiKeyProvider       auth.APIKeyProvider
	accessTokenProvider  auth.AccessTokenProvider
	maxResponseBodyBytes int64
	defaultPolicy        ExecutionPolicy
	classPolicies        map[traffic.Class]ExecutionPolicy
}

type requestCredentials struct {
	apiKey      string
	accessToken string
}

type apiKeyResultReporter interface {
	ReportAPIKeyResult(req *http.Request, key string, statusCode int, err error)
}

// RequestPreparer mutates one fully-built request before transport execution.
type RequestPreparer func(req *http.Request) error

// RetryBackoffConfig controls local retry delay behavior.
type RetryBackoffConfig struct {
	BaseDelay         time.Duration
	MaxDelay          time.Duration
	RespectRetryAfter bool
}

// ExecutionPolicy routes one request to one transport and retry profile.
type ExecutionPolicy struct {
	Retry          int
	RetryBackoff   RetryBackoffConfig
	Transport      Transport
	CacheRuntime   CacheRuntime
	BlockRuntime   BlockRuntime
	PrepareRequest RequestPreparer
}

// DefaultRetryBackoffConfig returns the SDK retry defaults.
func DefaultRetryBackoffConfig() RetryBackoffConfig {
	return RetryBackoffConfig{
		BaseDelay:         100 * time.Millisecond,
		MaxDelay:          2 * time.Second,
		RespectRetryAfter: true,
	}
}

// NewExecutor creates a request executor.
func NewExecutor(baseURL string, apiKeyProvider auth.APIKeyProvider, accessTokenProvider auth.AccessTokenProvider, maxResponseBodyBytes int64, defaultPolicy ExecutionPolicy, classPolicies map[traffic.Class]ExecutionPolicy) (*Executor, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "invalid base url", nil, err)
	}
	if maxResponseBodyBytes <= 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max response body bytes must be greater than zero", nil, nil)
	}
	if err := validateExecutionPolicy(defaultPolicy); err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, err.Error(), nil, nil)
	}
	resolvedClassPolicies := make(map[traffic.Class]ExecutionPolicy, len(classPolicies))
	for class, policy := range classPolicies {
		class = traffic.NormalizeClass(class)
		if err := validateExecutionPolicy(policy); err != nil {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, err.Error(), nil, nil)
		}
		resolvedClassPolicies[class] = policy
	}
	return &Executor{
		baseURL:              parsed,
		apiKeyProvider:       apiKeyProvider,
		accessTokenProvider:  accessTokenProvider,
		maxResponseBodyBytes: maxResponseBodyBytes,
		defaultPolicy:        defaultPolicy,
		classPolicies:        resolvedClassPolicies,
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
	class := e.executionClass(ctx, spec.TrafficClass)
	policy := e.executionPolicyForClass(class)

	var lastErr error
	for attempt := 0; attempt <= policy.Retry; attempt++ {
		req, err := e.buildRequest(ctx, resolved, spec, creds)
		if err != nil {
			return nil, err
		}
		req = req.WithContext(traffic.WithClass(req.Context(), class))
		if policy.BlockRuntime != nil {
			req = req.WithContext(traffic.WithBlockDetection(req.Context()))
		}
		if policy.PrepareRequest != nil {
			if err := policy.PrepareRequest(req); err != nil {
				return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "prepare request failed", nil, err)
			}
		}
		cacheLookup := cacheLookup{}
		if policy.CacheRuntime != nil && requestCacheable(req) {
			cacheLookup = policy.CacheRuntime.lookup(req, time.Now())
			if cacheLookup.fresh {
				return cloneBytes(cacheLookup.body), nil
			}
			if cacheLookupAllowsConditionalRequest(cacheLookup) {
				applyConditionalCacheHeaders(req, cacheLookup)
			}
		}

		resp, err := policy.Transport.Do(ctx, req)
		if err != nil {
			lastErr = sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, err)
			if shouldRetryTransport(ctx, err) && attempt < policy.Retry {
				if !sleepBeforeRetry(ctx, attempt, nil, policy.RetryBackoff) {
					return nil, sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
				}
				continue
			}
			return nil, lastErr
		}
		if resp.StatusCode == http.StatusNotModified && cacheLookup.found && policy.CacheRuntime != nil {
			if body, ok := policy.CacheRuntime.refresh(cacheLookup, resp, time.Now()); ok {
				_ = resp.Body.Close()
				return body, nil
			}
		}

		body, readErr := readAndCloseBody(resp, e.maxResponseBodyBytes)
		if readErr != nil {
			lastErr = sdkerrors.New(sdkerrors.KindTransport, resp.StatusCode, "read response body failed", nil, readErr)
			if attempt < policy.Retry {
				if !sleepBeforeRetry(ctx, attempt, resp, policy.RetryBackoff) {
					return nil, sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
				}
				continue
			}
			return nil, lastErr
		}
		e.reportAPIKeyResult(req, creds, resp.StatusCode, nil)
		if policy.BlockRuntime != nil {
			if blockResult := policy.BlockRuntime.detect(req, resp, body); blockResult != nil {
				lastErr = sdkerrors.New(blockResult.ErrorKind, resp.StatusCode, blockResult.Message, body, nil)
				if blockResult.Retryable && attempt < policy.Retry {
					if !sleepBeforeRetry(ctx, attempt, resp, policy.RetryBackoff) {
						return nil, sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
					}
					continue
				}
				return nil, lastErr
			}
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			lastErr = sdkerrors.New(
				sdkerrors.KindHTTPStatus,
				resp.StatusCode,
				fmt.Sprintf("unexpected http status %d", resp.StatusCode),
				body,
				nil,
			)
			if shouldRetryStatus(resp.StatusCode, e.apiKeyProvider != nil, e.accessTokenProvider != nil) && attempt < policy.Retry {
				creds, err = e.rotateRetryCredentials(req, resp.StatusCode, creds)
				if err != nil {
					return nil, err
				}
				if !sleepBeforeRetry(ctx, attempt, resp, policy.RetryBackoff) {
					return nil, sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
				}
				continue
			}
			return nil, lastErr
		}
		if policy.CacheRuntime != nil && requestCacheable(req) {
			policy.CacheRuntime.store(req, resp, body, time.Now())
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

func (e *Executor) rotateRetryCredentials(req *http.Request, statusCode int, creds requestCredentials) (requestCredentials, error) {
	if !shouldRotateCredentials(statusCode) {
		return creds, nil
	}

	var err error
	if e.apiKeyProvider != nil {
		creds.apiKey, err = e.resolveAPIKey(req)
		if err != nil {
			return requestCredentials{}, err
		}
	}
	if e.accessTokenProvider != nil {
		creds.accessToken, err = e.resolveAccessToken(req)
		if err != nil {
			return requestCredentials{}, err
		}
	}

	return creds, nil
}

func (e *Executor) reportAPIKeyResult(req *http.Request, creds requestCredentials, statusCode int, err error) {
	if e.apiKeyProvider == nil || creds.apiKey == "" {
		return
	}
	reporter, ok := e.apiKeyProvider.(apiKeyResultReporter)
	if !ok {
		return
	}
	reporter.ReportAPIKeyResult(req, creds.apiKey, statusCode, err)
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

func (e *Executor) executionClass(ctx context.Context, specClass traffic.Class) traffic.Class {
	class := traffic.NormalizeClass(specClass)
	if ctxClass, ok := traffic.ClassFromContext(ctx); ok {
		class = ctxClass
	}
	return class
}

func (e *Executor) executionPolicyForClass(class traffic.Class) ExecutionPolicy {
	if policy, ok := e.classPolicies[class]; ok {
		return policy
	}
	return e.defaultPolicy
}

func validateExecutionPolicy(policy ExecutionPolicy) error {
	if policy.Transport == nil {
		return fmt.Errorf("transport is required")
	}
	if policy.Retry < 0 {
		return fmt.Errorf("retry must not be negative")
	}
	if policy.RetryBackoff.BaseDelay <= 0 {
		return fmt.Errorf("retry base delay must be greater than zero")
	}
	if policy.RetryBackoff.MaxDelay <= 0 {
		return fmt.Errorf("retry max delay must be greater than zero")
	}
	if policy.RetryBackoff.MaxDelay < policy.RetryBackoff.BaseDelay {
		return fmt.Errorf("retry max delay must be greater than or equal to base delay")
	}
	return nil
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

func shouldRetryStatus(statusCode int, hasAPIKeyProvider, hasAccessTokenProvider bool) bool {
	if statusCode >= http.StatusInternalServerError {
		return true
	}
	if !hasAPIKeyProvider && !hasAccessTokenProvider {
		return false
	}
	return statusCode == http.StatusUnauthorized || statusCode == http.StatusTooManyRequests
}

func shouldRotateCredentials(statusCode int) bool {
	return statusCode == http.StatusUnauthorized || statusCode == http.StatusTooManyRequests
}

func sleepBeforeRetry(ctx context.Context, attempt int, resp *http.Response, cfg RetryBackoffConfig) bool {
	delay := retryDelay(attempt, resp, time.Now(), cfg)
	select {
	case <-ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}

func retryDelay(attempt int, resp *http.Response, now time.Time, cfg RetryBackoffConfig) time.Duration {
	if cfg.RespectRetryAfter {
		if delay, ok := retryAfterDelay(resp, now); ok {
			return delay
		}
	}

	delay := cfg.BaseDelay
	for i := 0; i < attempt; i++ {
		if delay >= cfg.MaxDelay {
			break
		}
		if delay > cfg.MaxDelay/2 {
			delay = cfg.MaxDelay
			break
		}
		delay *= 2
	}
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}

	jitterWindow := delay / 2
	if jitterWindow <= 0 {
		return delay
	}
	return delay + time.Duration(rand.Int64N(int64(jitterWindow)+1))
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
