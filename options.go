package steam

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/GoFurry/steam-go/internal/auth"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/transport"
	"golang.org/x/time/rate"
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
	apiKeyProvider       APIKeyProvider
	accessTokenProvider  AccessTokenProvider
	baseURL              string
	httpClient           *http.Client
	timeout              time.Duration
	retry                int
	rateLimiter          transport.RateLimiterConfig
	retryBackoff         request.RetryBackoffConfig
	cookieJar            http.CookieJar
	cookieJarConfigured  bool
	maxResponseBodyBytes int64
	proxySelector        ProxySelector
}

func defaultClientConfig() clientConfig {
	return clientConfig{
		baseURL:              defaultBaseURL,
		timeout:              10 * time.Second,
		retry:                0,
		retryBackoff:         request.DefaultRetryBackoffConfig(),
		maxResponseBodyBytes: 16 << 20,
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

// WithCookieJar injects a caller-managed cookie jar.
// Passing nil explicitly disables cookie persistence.
func WithCookieJar(jar http.CookieJar) Option {
	return func(cfg *clientConfig) error {
		cfg.cookieJar = jar
		cfg.cookieJarConfigured = true
		return nil
	}
}

// WithDefaultCookieJar enables session persistence with the standard library cookie jar.
func WithDefaultCookieJar() Option {
	return func(cfg *clientConfig) error {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return fmt.Errorf("create default cookie jar: %w", err)
		}
		cfg.cookieJar = jar
		cfg.cookieJarConfigured = true
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

// WithRetryBackoff configures the local retry backoff behavior.
func WithRetryBackoff(baseDelay, maxDelay time.Duration) Option {
	return func(cfg *clientConfig) error {
		if baseDelay <= 0 {
			return fmt.Errorf("retry base delay must be greater than zero")
		}
		if maxDelay <= 0 {
			return fmt.Errorf("retry max delay must be greater than zero")
		}
		if maxDelay < baseDelay {
			return fmt.Errorf("retry max delay must be greater than or equal to base delay")
		}
		cfg.retryBackoff.BaseDelay = baseDelay
		cfg.retryBackoff.MaxDelay = maxDelay
		return nil
	}
}

// WithRetryRespectRetryAfter controls whether Retry-After should override local backoff.
func WithRetryRespectRetryAfter(enabled bool) Option {
	return func(cfg *clientConfig) error {
		cfg.retryBackoff.RespectRetryAfter = enabled
		return nil
	}
}

// WithRateLimit enables a simple requests-per-second limiter.
func WithRateLimit(requestsPerSecond int) Option {
	return func(cfg *clientConfig) error {
		if requestsPerSecond < 0 {
			return fmt.Errorf("rate limit must not be negative")
		}
		if requestsPerSecond == 0 {
			cfg.rateLimiter = transport.RateLimiterConfig{}
			return nil
		}
		cfg.rateLimiter = transport.RateLimiterConfig{
			Limit: rate.Limit(requestsPerSecond),
			Burst: requestsPerSecond,
		}
		return nil
	}
}

// WithRateLimiter enables an explicit token-bucket limiter.
//
// Pass zero values for both limit and burst to disable rate limiting.
func WithRateLimiter(limit rate.Limit, burst int) Option {
	return func(cfg *clientConfig) error {
		if limit < 0 {
			return fmt.Errorf("rate limit must not be negative")
		}
		if burst < 0 {
			return fmt.Errorf("rate limiter burst must not be negative")
		}
		if limit == 0 || burst == 0 {
			if limit == 0 && burst == 0 {
				cfg.rateLimiter = transport.RateLimiterConfig{}
				return nil
			}
			return fmt.Errorf("rate limiter limit and burst must both be zero to disable")
		}
		cfg.rateLimiter = transport.RateLimiterConfig{
			Limit: limit,
			Burst: burst,
		}
		return nil
	}
}

// WithMaxResponseBodyBytes limits how many bytes the SDK will buffer for one response body.
func WithMaxResponseBodyBytes(max int64) Option {
	return func(cfg *clientConfig) error {
		if max <= 0 {
			return fmt.Errorf("max response body bytes must be greater than zero")
		}
		cfg.maxResponseBodyBytes = max
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
