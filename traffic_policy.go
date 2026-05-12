package steam

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/GoFurry/steam-go/internal/request"
	itraffic "github.com/GoFurry/steam-go/internal/traffic"
	"golang.org/x/time/rate"
)

// TrafficClass identifies one request traffic category.
type TrafficClass = itraffic.Class

const (
	TrafficClassOfficialAPI     TrafficClass = itraffic.ClassOfficialAPI
	TrafficClassPublicStorePage TrafficClass = itraffic.ClassPublicStorePage
)

// RetryBackoffConfig exposes the SDK retry backoff shape for policy overrides.
type RetryBackoffConfig = request.RetryBackoffConfig

// DefaultRetryBackoffConfig returns the SDK retry backoff defaults.
func DefaultRetryBackoffConfig() RetryBackoffConfig {
	return request.DefaultRetryBackoffConfig()
}

// TrafficRateLimiterPolicy overrides per-class token-bucket settings.
type TrafficRateLimiterPolicy struct {
	Limit rate.Limit
	Burst int
}

// TrafficRetryPolicy overrides per-class retry behavior.
type TrafficRetryPolicy struct {
	Retry   int
	Backoff RetryBackoffConfig
}

// TrafficHostControlPolicy overrides per-class host-scoped request controls.
type TrafficHostControlPolicy struct {
	RateLimiter   *TrafficRateLimiterPolicy
	MaxConcurrent int
}

// TrafficSessionControlPolicy overrides per-class session-scoped request controls.
type TrafficSessionControlPolicy struct {
	RateLimiter   *TrafficRateLimiterPolicy
	MaxConcurrent int
}

// TrafficCachePolicy overrides per-class conditional-request and short-cache behavior.
type TrafficCachePolicy struct {
	TTL time.Duration
}

// TrafficBlockPolicy overrides per-class block detection behavior.
type TrafficBlockPolicy struct {
	HTMLSniffBytes int
}

// TransportHook customizes one class-specific HTTP execution stack after the SDK assembles its base client.
type TransportHook interface {
	WrapHTTPClient(class TrafficClass, base *http.Client) (*http.Client, error)
}

// ProxyAwareTransportHook customizes one class-specific HTTP client before proxy selection is wrapped.
//
// The hook receives the current traffic class, one assembled base client, and the class-specific proxy selector.
// It is responsible for returning the final http.Client that should be used for this traffic class.
type ProxyAwareTransportHook interface {
	TransportHook
	WrapHTTPClientWithProxy(class TrafficClass, base *http.Client, selector ProxySelector) (*http.Client, error)
}

// TransportHookFunc adapts one function into a TransportHook.
type TransportHookFunc func(class TrafficClass, base *http.Client) (*http.Client, error)

// WrapHTTPClient implements TransportHook.
func (f TransportHookFunc) WrapHTTPClient(class TrafficClass, base *http.Client) (*http.Client, error) {
	return f(class, base)
}

// ProxyAwareTransportHookFunc adapts one function into a ProxyAwareTransportHook.
type ProxyAwareTransportHookFunc func(class TrafficClass, base *http.Client, selector ProxySelector) (*http.Client, error)

// WrapHTTPClientWithProxy implements ProxyAwareTransportHook.
func (f ProxyAwareTransportHookFunc) WrapHTTPClientWithProxy(class TrafficClass, base *http.Client, selector ProxySelector) (*http.Client, error) {
	return f(class, base, selector)
}

// WrapHTTPClient implements TransportHook.
//
// When used through the narrower TransportHook interface, no proxy selector is provided.
func (f ProxyAwareTransportHookFunc) WrapHTTPClient(class TrafficClass, base *http.Client) (*http.Client, error) {
	return f(class, base, nil)
}

// TrafficPolicy overrides selected request behavior for one traffic class.
type TrafficPolicy struct {
	ProxySelector   ProxySelector
	CookieJar       http.CookieJar
	RateLimiter     *TrafficRateLimiterPolicy
	Retry           *TrafficRetryPolicy
	HostControl     *TrafficHostControlPolicy
	SessionControl  *TrafficSessionControlPolicy
	Cache           *TrafficCachePolicy
	BlockPolicy     *TrafficBlockPolicy
	HeaderProfile   *HeaderProfile
	RefererSelector RefererSelector
	TransportHook   TransportHook
}

// WithTrafficClass attaches one traffic class to a request context.
func WithTrafficClass(ctx context.Context, class TrafficClass) context.Context {
	return itraffic.WithClass(ctx, class)
}

// WithRequestSessionKey attaches one explicit request-session key to a context.
func WithRequestSessionKey(ctx context.Context, key string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return ctx
	}
	return itraffic.WithRequestSessionKey(ctx, key)
}

type trafficPolicyConfig struct {
	proxySelector     ProxySelector
	cookieJar         http.CookieJar
	rateLimiter       *TrafficRateLimiterPolicy
	retry             *TrafficRetryPolicy
	hostControl       *TrafficHostControlPolicy
	sessionControl    *TrafficSessionControlPolicy
	cache             *TrafficCachePolicy
	blockPolicy       *TrafficBlockPolicy
	headerProfile     *HeaderProfile
	refererSelector   RefererSelector
	transportHook     TransportHook
	cookieJarProvided bool
}

// WithTrafficPolicy configures one per-class request policy override.
func WithTrafficPolicy(class TrafficClass, policy TrafficPolicy) Option {
	return func(cfg *clientConfig) error {
		if !supportedTrafficClass(class) {
			return fmt.Errorf("unsupported traffic class")
		}
		class = normalizeTrafficClass(class)
		if err := validateTrafficRateLimiterPolicy(policy.RateLimiter); err != nil {
			return err
		}
		if policy.Retry != nil {
			if policy.Retry.Retry < 0 {
				return fmt.Errorf("traffic policy retry must not be negative")
			}
			if policy.Retry.Backoff.BaseDelay <= 0 {
				return fmt.Errorf("traffic policy retry base delay must be greater than zero")
			}
			if policy.Retry.Backoff.MaxDelay <= 0 {
				return fmt.Errorf("traffic policy retry max delay must be greater than zero")
			}
			if policy.Retry.Backoff.MaxDelay < policy.Retry.Backoff.BaseDelay {
				return fmt.Errorf("traffic policy retry max delay must be greater than or equal to base delay")
			}
		}
		if err := validateTrafficHostControlPolicy(policy.HostControl); err != nil {
			return err
		}
		if err := validateTrafficSessionControlPolicy(policy.SessionControl); err != nil {
			return err
		}
		if err := validateTrafficCachePolicy(policy.Cache); err != nil {
			return err
		}
		if err := validateTrafficBlockPolicy(class, policy.BlockPolicy); err != nil {
			return err
		}
		if cfg.trafficPolicies == nil {
			cfg.trafficPolicies = make(map[TrafficClass]trafficPolicyConfig)
		}
		cfg.trafficPolicies[class] = trafficPolicyConfig{
			proxySelector:     policy.ProxySelector,
			cookieJar:         policy.CookieJar,
			rateLimiter:       policy.RateLimiter,
			retry:             policy.Retry,
			hostControl:       policy.HostControl,
			sessionControl:    policy.SessionControl,
			cache:             policy.Cache,
			blockPolicy:       policy.BlockPolicy,
			headerProfile:     cloneHeaderProfile(policy.HeaderProfile),
			refererSelector:   policy.RefererSelector,
			transportHook:     policy.TransportHook,
			cookieJarProvided: policy.CookieJar != nil,
		}
		return nil
	}
}

func normalizeTrafficClass(class TrafficClass) TrafficClass {
	return itraffic.NormalizeClass(class)
}

func supportedTrafficClass(class TrafficClass) bool {
	switch class {
	case TrafficClassOfficialAPI, TrafficClassPublicStorePage:
		return true
	default:
		return false
	}
}

func validateTrafficRateLimiterPolicy(policy *TrafficRateLimiterPolicy) error {
	if policy == nil {
		return nil
	}
	if policy.Limit < 0 {
		return fmt.Errorf("traffic policy rate limit must not be negative")
	}
	if policy.Burst < 0 {
		return fmt.Errorf("traffic policy rate limiter burst must not be negative")
	}
	if policy.Limit == 0 || policy.Burst == 0 {
		if policy.Limit != 0 || policy.Burst != 0 {
			return fmt.Errorf("traffic policy rate limiter limit and burst must both be zero to disable")
		}
	}
	return nil
}

func validateTrafficHostControlPolicy(policy *TrafficHostControlPolicy) error {
	if policy == nil {
		return nil
	}
	if policy.MaxConcurrent < 0 {
		return fmt.Errorf("traffic policy host max concurrent must not be negative")
	}
	if err := validateTrafficRateLimiterPolicy(policy.RateLimiter); err != nil {
		return fmt.Errorf("traffic policy host control: %w", err)
	}
	return nil
}

func validateTrafficSessionControlPolicy(policy *TrafficSessionControlPolicy) error {
	if policy == nil {
		return nil
	}
	if policy.MaxConcurrent < 0 {
		return fmt.Errorf("traffic policy session max concurrent must not be negative")
	}
	if err := validateTrafficRateLimiterPolicy(policy.RateLimiter); err != nil {
		return fmt.Errorf("traffic policy session control: %w", err)
	}
	return nil
}

func validateTrafficCachePolicy(policy *TrafficCachePolicy) error {
	if policy == nil {
		return nil
	}
	if policy.TTL <= 0 {
		return fmt.Errorf("traffic policy cache ttl must be greater than zero")
	}
	return nil
}

func validateTrafficBlockPolicy(class TrafficClass, policy *TrafficBlockPolicy) error {
	if policy == nil {
		return nil
	}
	if normalizeTrafficClass(class) != TrafficClassPublicStorePage {
		return fmt.Errorf("traffic policy block detection is only supported for public store page traffic")
	}
	if policy.HTMLSniffBytes < 0 {
		return fmt.Errorf("traffic policy block html sniff bytes must not be negative")
	}
	return nil
}
