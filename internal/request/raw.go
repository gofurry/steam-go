package request

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/traffic"
)

// ExecuteRawHTTPRequest executes one already-built absolute HTTP request with one execution policy.
func ExecuteRawHTTPRequest(ctx context.Context, req *http.Request, maxResponseBodyBytes int64, policy ExecutionPolicy, retryableOverride *bool) (HTTPResult, error) {
	if req == nil {
		return HTTPResult{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "request is required", nil, nil)
	}
	if req.URL == nil || req.URL.Scheme == "" || req.URL.Host == "" {
		return HTTPResult{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "request url must be absolute", nil, nil)
	}
	if maxResponseBodyBytes <= 0 {
		return HTTPResult{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max response body bytes must be greater than zero", nil, nil)
	}
	if err := validateExecutionPolicy(policy); err != nil {
		return HTTPResult{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, err.Error(), nil, nil)
	}

	baseReq, err := freezeRequestForRetries(req)
	if err != nil {
		return HTTPResult{}, err
	}
	class := traffic.NormalizeClass(traffic.ClassOfficialAPI)
	if ctxClass, ok := traffic.ClassFromContext(ctx); ok {
		class = traffic.NormalizeClass(ctxClass)
	}
	started := time.Now()
	var lastReq *http.Request
	attempts := 0
	retryable := requestRetryable(req.Method, retryableOverride)

	var lastErr error
	var lastResult HTTPResult
	for attempt := 0; attempt <= policy.Retry; attempt++ {
		attempts = attempt + 1
		execReq, err := cloneRequestForExecution(baseReq, ctx)
		if err != nil {
			observeRequest(policy.Observer, lastReq, class, 0, err, attempts, false, false, started)
			return HTTPResult{}, err
		}
		lastReq = execReq
		if policy.BlockRuntime != nil {
			execReq = execReq.WithContext(traffic.WithBlockDetection(execReq.Context()))
		}
		if policy.PrepareRequest != nil {
			if err := policy.PrepareRequest(execReq); err != nil {
				observedErr := sdkerrors.New(sdkerrors.KindRequestBuild, 0, "prepare request failed", nil, err)
				observeRequest(policy.Observer, execReq, class, 0, observedErr, attempts, false, false, started)
				return HTTPResult{}, observedErr
			}
		}

		cacheLookup := cacheLookup{}
		if policy.CacheRuntime != nil && requestCacheable(execReq) {
			cacheLookup = policy.CacheRuntime.lookup(execReq, time.Now())
			if cacheLookup.fresh {
				observeRequest(policy.Observer, execReq, class, cacheLookup.result.StatusCode, nil, attempts, true, false, started)
				return cloneHTTPResult(cacheLookup.result), nil
			}
			if cacheLookupAllowsConditionalRequest(cacheLookup) {
				applyConditionalCacheHeaders(execReq, cacheLookup)
			}
		}

		resp, err := policy.Transport.Do(execReq.Context(), execReq)
		if err != nil {
			lastErr = sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, err)
			if retryable && shouldRetryTransport(ctx, err) && attempt < policy.Retry {
				if !sleepBeforeRetry(ctx, attempt, nil, policy.RetryBackoff) {
					lastErr = sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
					observeRequest(policy.Observer, execReq, class, 0, lastErr, attempts, false, false, started)
					return HTTPResult{}, lastErr
				}
				continue
			}
			observeRequest(policy.Observer, execReq, class, 0, lastErr, attempts, false, false, started)
			return HTTPResult{}, lastErr
		}

		if resp.StatusCode == http.StatusNotModified && cacheLookup.found && policy.CacheRuntime != nil {
			if result, ok := policy.CacheRuntime.refresh(cacheLookup, resp, time.Now()); ok {
				_ = resp.Body.Close()
				return result, nil
			}
		}

		body, readErr := readAndCloseBody(resp, maxResponseBodyBytes)
		if readErr != nil {
			lastErr = sdkerrors.New(sdkerrors.KindTransport, resp.StatusCode, "read response body failed", nil, readErr)
			if retryable && attempt < policy.Retry {
				if !sleepBeforeRetry(ctx, attempt, resp, policy.RetryBackoff) {
					lastErr = sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
					observeRequest(policy.Observer, execReq, class, resp.StatusCode, lastErr, attempts, false, false, started)
					return HTTPResult{}, lastErr
				}
				continue
			}
			observeRequest(policy.Observer, execReq, class, resp.StatusCode, lastErr, attempts, false, false, started)
			return HTTPResult{}, lastErr
		}

		result := HTTPResult{
			StatusCode: resp.StatusCode,
			Header:     cloneHeader(resp.Header),
			Body:       body,
		}
		if resp.Request != nil && resp.Request.URL != nil {
			result.FinalURL = cloneURL(resp.Request.URL)
		}

		if policy.BlockRuntime != nil {
			if blockResult := policy.BlockRuntime.detect(execReq, resp, body); blockResult != nil {
				result.Block = &BlockResult{
					ErrorKind: blockResult.ErrorKind,
					Message:   blockResult.Message,
					Retryable: blockResult.Retryable,
				}
				if retryable && blockResult.Retryable && attempt < policy.Retry {
					lastResult = cloneHTTPResult(result)
					if !sleepBeforeRetry(ctx, attempt, resp, policy.RetryBackoff) {
						lastErr = sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
						observeRequest(policy.Observer, execReq, class, resp.StatusCode, lastErr, attempts, false, true, started)
						return HTTPResult{}, lastErr
					}
					continue
				}
				observeRequest(policy.Observer, execReq, class, resp.StatusCode, nil, attempts, false, true, started)
				return result, nil
			}
		}

		if retryable && resp.StatusCode >= http.StatusInternalServerError && attempt < policy.Retry {
			lastResult = cloneHTTPResult(result)
			if !sleepBeforeRetry(ctx, attempt, resp, policy.RetryBackoff) {
				lastErr = sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, ctx.Err())
				observeRequest(policy.Observer, execReq, class, resp.StatusCode, lastErr, attempts, false, false, started)
				return HTTPResult{}, lastErr
			}
			continue
		}

		if policy.CacheRuntime != nil && requestCacheable(execReq) && resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			policy.CacheRuntime.store(execReq, resp, result, time.Now())
		}

		observeRequest(policy.Observer, execReq, class, resp.StatusCode, nil, attempts, false, false, started)
		return result, nil
	}

	if lastErr != nil {
		return HTTPResult{}, lastErr
	}
	if lastResult.StatusCode != 0 || len(lastResult.Body) > 0 || lastResult.FinalURL != nil || lastResult.Block != nil {
		observeRequest(policy.Observer, lastReq, class, lastResult.StatusCode, nil, attempts, false, lastResult.Block != nil, started)
		return lastResult, nil
	}
	lastErr = sdkerrors.New(sdkerrors.KindTransport, 0, "request execution failed", nil, nil)
	observeRequest(policy.Observer, lastReq, class, 0, lastErr, attempts, false, false, started)
	return HTTPResult{}, lastErr
}

func freezeRequestForRetries(req *http.Request) (*http.Request, error) {
	if req == nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "request is required", nil, nil)
	}
	if req.Body == nil {
		return req.Clone(req.Context()), nil
	}
	if req.GetBody != nil {
		cloned := req.Clone(req.Context())
		body, err := req.GetBody()
		if err != nil {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "clone request body failed", nil, err)
		}
		cloned.Body = body
		cloned.GetBody = req.GetBody
		return cloned, nil
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "read request body failed", nil, err)
	}
	_ = req.Body.Close()
	restoreBody := func() io.ReadCloser {
		return io.NopCloser(bytes.NewReader(bodyBytes))
	}
	req.Body = restoreBody()
	req.GetBody = func() (io.ReadCloser, error) {
		return restoreBody(), nil
	}

	cloned := req.Clone(req.Context())
	cloned.Body = restoreBody()
	cloned.GetBody = req.GetBody
	return cloned, nil
}

func cloneRequestForExecution(req *http.Request, ctx context.Context) (*http.Request, error) {
	if req == nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "request is required", nil, nil)
	}
	cloned := req.Clone(ctx)
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "clone request body failed", nil, err)
		}
		cloned.Body = body
		cloned.GetBody = req.GetBody
	}
	return cloned, nil
}
