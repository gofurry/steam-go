package steam

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/GoFurry/steam-go/internal/auth"
)

const defaultBaseURL = "https://api.steampowered.com"

// ProxySelector chooses a proxy URL for a request.
type ProxySelector interface {
	Next(req *http.Request) (*url.URL, error)
}

// APIKeyProvider resolves the API key used for each request.
type APIKeyProvider = auth.APIKeyProvider

// AccessTokenProvider resolves the access token used for each request.
type AccessTokenProvider = auth.AccessTokenProvider

// Option mutates client construction settings.
type Option func(*clientConfig) error

type clientConfig struct {
	apiKeyProvider      APIKeyProvider
	accessTokenProvider AccessTokenProvider
	baseURL             string
	httpClient          *http.Client
	timeout             time.Duration
	retry               int
	rateLimit           int
	proxySelector       ProxySelector
	logger              Logger
}

func defaultClientConfig() clientConfig {
	return clientConfig{
		baseURL:   defaultBaseURL,
		timeout:   10 * time.Second,
		retry:     0,
		rateLimit: 0,
		logger:    noopLogger{},
	}
}

// WithAPIKey configures a static Steam Web API key.
// An empty key disables key injection.
func WithAPIKey(key string) Option {
	return func(cfg *clientConfig) error {
		key = strings.TrimSpace(key)
		if key == "" {
			cfg.apiKeyProvider = nil
			return nil
		}
		cfg.apiKeyProvider = auth.NewStaticKeyProvider(key)
		return nil
	}
}

// WithAPIKeys configures a round-robin key provider.
// Empty keys are ignored; an empty final set disables key injection.
func WithAPIKeys(keys ...string) Option {
	return func(cfg *clientConfig) error {
		cfg.apiKeyProvider = auth.NewRoundRobinKeyProvider(keys...)
		return nil
	}
}

// WithAPIKeyProvider injects a custom per-request key provider.
func WithAPIKeyProvider(provider APIKeyProvider) Option {
	return func(cfg *clientConfig) error {
		cfg.apiKeyProvider = provider
		return nil
	}
}

// WithAccessToken configures a static Steam access token.
// An empty token disables access token injection.
func WithAccessToken(token string) Option {
	return func(cfg *clientConfig) error {
		token = strings.TrimSpace(token)
		if token == "" {
			cfg.accessTokenProvider = nil
			return nil
		}
		cfg.accessTokenProvider = auth.NewStaticAccessTokenProvider(token)
		return nil
	}
}

// WithAccessTokens configures a round-robin access token provider.
// Empty values are ignored; an empty final set disables token injection.
func WithAccessTokens(tokens ...string) Option {
	return func(cfg *clientConfig) error {
		cfg.accessTokenProvider = auth.NewRoundRobinAccessTokenProvider(tokens...)
		return nil
	}
}

// WithAccessTokenProvider injects a custom per-request access token provider.
func WithAccessTokenProvider(provider AccessTokenProvider) Option {
	return func(cfg *clientConfig) error {
		cfg.accessTokenProvider = provider
		return nil
	}
}

// WithBaseURL overrides the default Steam Web API base URL.
func WithBaseURL(rawURL string) Option {
	return func(cfg *clientConfig) error {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("parse base url: %w", err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("base url must include scheme and host")
		}
		cfg.baseURL = strings.TrimRight(parsed.String(), "/")
		return nil
	}
}

// WithHTTPClient injects a caller-managed HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) error {
		if client == nil {
			return fmt.Errorf("http client must not be nil")
		}
		cfg.httpClient = client
		return nil
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(cfg *clientConfig) error {
		if timeout <= 0 {
			return fmt.Errorf("timeout must be greater than zero")
		}
		cfg.timeout = timeout
		return nil
	}
}

// WithRetry sets the retry count for transport and 429/5xx responses.
func WithRetry(retry int) Option {
	return func(cfg *clientConfig) error {
		if retry < 0 {
			return fmt.Errorf("retry must not be negative")
		}
		cfg.retry = retry
		return nil
	}
}

// WithRateLimit enables a simple requests-per-second limiter.
func WithRateLimit(requestsPerSecond int) Option {
	return func(cfg *clientConfig) error {
		if requestsPerSecond < 0 {
			return fmt.Errorf("rate limit must not be negative")
		}
		cfg.rateLimit = requestsPerSecond
		return nil
	}
}

// WithProxySelector injects a proxy chooser.
func WithProxySelector(selector ProxySelector) Option {
	return func(cfg *clientConfig) error {
		cfg.proxySelector = selector
		return nil
	}
}

// WithLogger injects the SDK logger implementation.
func WithLogger(logger Logger) Option {
	return func(cfg *clientConfig) error {
		if logger == nil {
			return fmt.Errorf("logger must not be nil")
		}
		cfg.logger = logger
		return nil
	}
}
