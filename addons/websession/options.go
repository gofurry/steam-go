package websession

import (
	"net/http"
	"net/url"
	"time"
)

const (
	defaultLoginBaseURL         = "https://login.steampowered.com"
	defaultStoreBaseURL         = "https://store.steampowered.com"
	defaultCommunityBaseURL     = "https://steamcommunity.com"
	defaultTimeout              = 10 * time.Second
	defaultMaxResponseBodyBytes = 1 << 20
)

type Option func(*clientOptions) error

type clientOptions struct {
	httpClient           *http.Client
	httpClientConfigured bool
	timeout              time.Duration
	maxResponseBodyBytes int64
	loginBaseURL         *url.URL
	storeBaseURL         *url.URL
	communityBaseURL     *url.URL
}

func defaultClientOptions() (clientOptions, error) {
	loginBaseURL, err := parseAbsoluteURL(defaultLoginBaseURL)
	if err != nil {
		return clientOptions{}, err
	}
	storeBaseURL, err := parseAbsoluteURL(defaultStoreBaseURL)
	if err != nil {
		return clientOptions{}, err
	}
	communityBaseURL, err := parseAbsoluteURL(defaultCommunityBaseURL)
	if err != nil {
		return clientOptions{}, err
	}
	return clientOptions{
		httpClient:           http.DefaultClient,
		timeout:              defaultTimeout,
		maxResponseBodyBytes: defaultMaxResponseBodyBytes,
		loginBaseURL:         loginBaseURL,
		storeBaseURL:         storeBaseURL,
		communityBaseURL:     communityBaseURL,
	}, nil
}

func WithHTTPClient(client *http.Client) Option {
	return func(opts *clientOptions) error {
		if client == nil {
			return configError("with_http_client", "http client must not be nil", nil)
		}
		opts.httpClient = client
		opts.httpClientConfigured = true
		return nil
	}
}

func WithTimeout(d time.Duration) Option {
	return func(opts *clientOptions) error {
		if d <= 0 {
			return configError("with_timeout", "timeout must be greater than zero", nil)
		}
		opts.timeout = d
		return nil
	}
}

func WithMaxResponseBodyBytes(max int64) Option {
	return func(opts *clientOptions) error {
		if max <= 0 {
			return configError("with_max_response_body_bytes", "max response body bytes must be greater than zero", nil)
		}
		opts.maxResponseBodyBytes = max
		return nil
	}
}

func WithLoginBaseURL(rawURL string) Option {
	return func(opts *clientOptions) error {
		parsed, err := parseAbsoluteURL(rawURL)
		if err != nil {
			return configError("with_login_base_url", "invalid login base url", err)
		}
		opts.loginBaseURL = parsed
		return nil
	}
}

func WithStoreBaseURL(rawURL string) Option {
	return func(opts *clientOptions) error {
		parsed, err := parseAbsoluteURL(rawURL)
		if err != nil {
			return configError("with_store_base_url", "invalid store base url", err)
		}
		opts.storeBaseURL = parsed
		return nil
	}
}

func WithCommunityBaseURL(rawURL string) Option {
	return func(opts *clientOptions) error {
		parsed, err := parseAbsoluteURL(rawURL)
		if err != nil {
			return configError("with_community_base_url", "invalid community base url", err)
		}
		opts.communityBaseURL = parsed
		return nil
	}
}

func configError(op, message string, err error) *Error {
	return &Error{Code: ErrorCodeConfig, Op: op, Message: message, Err: err}
}
