package freeclaim

import (
	"net/http"
	"net/url"
	"time"
)

const (
	defaultStoreBaseURL         = "https://store.steampowered.com"
	defaultTimeout              = 10 * time.Second
	defaultMaxResponseBodyBytes = 1 << 20
)

type Option func(*clientOptions) error

type clientOptions struct {
	httpClient           *http.Client
	httpClientConfigured bool
	timeout              time.Duration
	maxResponseBodyBytes int64
	storeBaseURL         *url.URL
}

func defaultClientOptions() (clientOptions, error) {
	storeBaseURL, err := parseAbsoluteURL(defaultStoreBaseURL)
	if err != nil {
		return clientOptions{}, err
	}
	return clientOptions{
		httpClient:           http.DefaultClient,
		timeout:              defaultTimeout,
		maxResponseBodyBytes: defaultMaxResponseBodyBytes,
		storeBaseURL:         storeBaseURL,
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

func configError(op, message string, err error) *Error {
	return &Error{Code: ErrorCodeConfig, Op: op, Message: message, Err: err}
}
