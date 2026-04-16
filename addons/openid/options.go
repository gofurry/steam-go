package openid

import (
	"net/http"
	"net/url"
	"time"
)

const (
	defaultEndpoint   = "https://steamcommunity.com/openid/login"
	defaultStateParam = "state"
)

type Option func(*verifierOptions) error

type verifierOptions struct {
	endpoint   *url.URL
	httpClient *http.Client
	timeout    time.Duration
	stateParam string
}

func defaultVerifierOptions() verifierOptions {
	return verifierOptions{
		httpClient: http.DefaultClient,
		timeout:    10 * time.Second,
		stateParam: defaultStateParam,
	}
}

func WithEndpoint(rawURL string) Option {
	return func(opts *verifierOptions) error {
		parsed, err := parseAbsoluteURL(rawURL)
		if err != nil {
			return &Error{
				Code:    ErrorCodeConfig,
				Op:      "with_endpoint",
				Message: "invalid endpoint",
				Err:     err,
			}
		}
		opts.endpoint = parsed
		return nil
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(opts *verifierOptions) error {
		if client == nil {
			return &Error{
				Code:    ErrorCodeConfig,
				Op:      "with_http_client",
				Message: "http client must not be nil",
			}
		}
		opts.httpClient = client
		return nil
	}
}

func WithTimeout(d time.Duration) Option {
	return func(opts *verifierOptions) error {
		if d <= 0 {
			return &Error{
				Code:    ErrorCodeConfig,
				Op:      "with_timeout",
				Message: "timeout must be greater than zero",
			}
		}
		opts.timeout = d
		return nil
	}
}

func WithStateParam(name string) Option {
	return func(opts *verifierOptions) error {
		if name == "" {
			return &Error{
				Code:    ErrorCodeConfig,
				Op:      "with_state_param",
				Message: "state param must not be empty",
			}
		}
		opts.stateParam = name
		return nil
	}
}
