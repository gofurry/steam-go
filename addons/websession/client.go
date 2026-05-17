package websession

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/api/authenticationservice"
)

type Client struct {
	auth                 *authenticationservice.Service
	sdkClient            *steam.Client
	httpClient           *http.Client
	timeout              time.Duration
	maxResponseBodyBytes int64
	loginBaseURL         *url.URL
	storeBaseURL         *url.URL
	communityBaseURL     *url.URL
}

type httpResult struct {
	StatusCode int
	FinalURL   *url.URL
	Header     http.Header
	Body       []byte
	Block      *steam.RawHTTPBlockResult
}

func NewClient(auth *authenticationservice.Service, opts ...Option) (*Client, error) {
	if auth == nil {
		return nil, configError("new_client", "authentication service must not be nil", nil)
	}
	options, err := defaultClientOptions()
	if err != nil {
		return nil, configError("new_client", "invalid default options", err)
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&options); err != nil {
			return nil, err
		}
	}
	return newClientWithOptions(auth, nil, options), nil
}

func NewClientFromSteamClient(client *steam.Client, opts ...Option) (*Client, error) {
	if client == nil {
		return nil, configError("new_client_from_steam_client", "steam client must not be nil", nil)
	}
	if client.API == nil || client.API.AuthenticationService == nil {
		return nil, configError("new_client_from_steam_client", "steam client authentication service must not be nil", nil)
	}
	options, err := defaultClientOptions()
	if err != nil {
		return nil, configError("new_client_from_steam_client", "invalid default options", err)
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&options); err != nil {
			return nil, err
		}
	}
	if options.httpClientConfigured {
		return nil, configError("new_client_from_steam_client", "with_http_client is not supported with NewClientFromSteamClient", nil)
	}
	return newClientWithOptions(client.API.AuthenticationService, client, options), nil
}

func newClientWithOptions(auth *authenticationservice.Service, sdkClient *steam.Client, options clientOptions) *Client {
	return &Client{
		auth:                 auth,
		sdkClient:            sdkClient,
		httpClient:           options.httpClient,
		timeout:              options.timeout,
		maxResponseBodyBytes: options.maxResponseBodyBytes,
		loginBaseURL:         cloneURL(options.loginBaseURL),
		storeBaseURL:         cloneURL(options.storeBaseURL),
		communityBaseURL:     cloneURL(options.communityBaseURL),
	}
}

func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, c.timeout)
}

func readBodyLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(r)
	}
	reader := &io.LimitedReader{R: r, N: maxBytes + 1}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("response body exceeds limit of %d bytes", maxBytes)
	}
	return body, nil
}

func parseAbsoluteURL(raw string) (*url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("url must not be empty")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("url must be absolute")
	}
	return parsed, nil
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	cloned := *u
	return &cloned
}

func resolveURL(base *url.URL, path string) *url.URL {
	if base == nil {
		return &url.URL{Path: path}
	}
	ref := &url.URL{Path: path}
	return cloneURL(base).ResolveReference(ref)
}

func (c *Client) doRequestWithJar(ctx context.Context, jar http.CookieJar, trafficClass steam.TrafficClass, method, rawURL string, body io.Reader, contentType string, headers map[string]string, op string) (httpResult, error) {
	if c.sdkClient != nil {
		return c.doSDKRequestWithJar(ctx, jar, trafficClass, method, rawURL, body, contentType, headers, op)
	}
	return c.doManualRequestWithJar(ctx, jar, method, rawURL, body, contentType, headers, op)
}

func (c *Client) doSDKRequestWithJar(ctx context.Context, jar http.CookieJar, trafficClass steam.TrafficClass, method, rawURL string, body io.Reader, contentType string, headers map[string]string, op string) (httpResult, error) {
	reqCtx, cancel := c.withTimeout(ctx)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, method, rawURL, body)
	if err != nil {
		return httpResult{}, &Error{Code: ErrorCodeRequestBuild, Op: op, Message: "build request failed", Err: err}
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	result, err := c.sdkClient.DoRawHTTPRequest(reqCtx, req, &steam.RawHTTPRequestOptions{
		TrafficClass:         trafficClass,
		CookieJar:            jar,
		MaxResponseBodyBytes: c.maxResponseBodyBytes,
	})
	if err != nil {
		return httpResult{}, wrapSDKRawError(op, err)
	}
	return httpResult{
		StatusCode: result.StatusCode,
		FinalURL:   cloneURL(result.FinalURL),
		Header:     result.Header.Clone(),
		Body:       append([]byte(nil), result.Body...),
		Block:      result.Block,
	}, nil
}

func (c *Client) doManualRequestWithJar(ctx context.Context, jar http.CookieJar, method, rawURL string, body io.Reader, contentType string, headers map[string]string, op string) (httpResult, error) {
	reqCtx, cancel := c.withTimeout(ctx)
	defer cancel()

	client := cloneHTTPClient(c.httpClient)
	client.Jar = jar

	req, err := http.NewRequestWithContext(reqCtx, method, rawURL, body)
	if err != nil {
		return httpResult{}, &Error{Code: ErrorCodeRequestBuild, Op: op, Message: "build request failed", Err: err}
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return httpResult{}, &Error{Code: ErrorCodeTransport, Op: op, Message: "request failed", Err: err}
	}
	defer resp.Body.Close()

	responseBody, readErr := readBodyLimited(resp.Body, c.maxResponseBodyBytes)
	if readErr != nil {
		return httpResult{}, &Error{Code: ErrorCodeTransport, Op: op, Message: "read response failed", Err: readErr}
	}

	result := httpResult{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       responseBody,
	}
	if resp.Request != nil && resp.Request.URL != nil {
		result.FinalURL = cloneURL(resp.Request.URL)
	}
	return result, nil
}

func wrapSDKRawError(op string, err error) error {
	var apiErr *steam.APIError
	if !errors.As(err, &apiErr) {
		return &Error{Code: ErrorCodeTransport, Op: op, Message: "request failed", Err: err}
	}
	switch apiErr.Kind {
	case steam.ErrorKindRequestBuild:
		return &Error{Code: ErrorCodeRequestBuild, Op: op, Message: apiErr.Message, Err: apiErr.Err}
	case steam.ErrorKindTransport:
		return &Error{Code: ErrorCodeTransport, Op: op, Message: apiErr.Message, Err: apiErr.Err}
	case steam.ErrorKindDecode:
		return &Error{Code: ErrorCodeDecode, Op: op, Message: apiErr.Message, Err: apiErr.Err}
	case steam.ErrorKindHTTPStatus:
		return &Error{Code: ErrorCodeHTTPStatus, Op: op, Message: apiErr.Message, Err: apiErr.Err}
	default:
		return &Error{Code: ErrorCodeVerify, Op: op, Message: apiErr.Message, Err: apiErr.Err}
	}
}
