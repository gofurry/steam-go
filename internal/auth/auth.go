package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// APIKeyProvider resolves a key for an outgoing request.
type APIKeyProvider interface {
	Next(req *http.Request) (string, error)
}

// AccessTokenProvider resolves an access token for an outgoing request.
type AccessTokenProvider interface {
	Next(req *http.Request) (string, error)
}

// KeyHealthConfig configures cooldown behavior for rotating API keys.
type KeyHealthConfig struct {
	FailureThreshold int
	Cooldown         time.Duration
}

// NewStaticKeyProvider creates a provider that always returns the same key.
func NewStaticKeyProvider(key string) APIKeyProvider {
	provider := newRoundRobinValuesProvider(key)
	if provider == nil {
		return nil
	}
	return provider
}

// NewRoundRobinKeyProvider creates a provider that rotates across keys.
// Empty values are ignored. If no keys remain, nil is returned.
func NewRoundRobinKeyProvider(keys ...string) APIKeyProvider {
	return newRoundRobinValuesProvider(keys...)
}

// NewHealthCheckedRoundRobinKeyProvider creates one rotating key provider that temporarily cools down keys
// after repeated 401/429 responses. Empty values are ignored; an empty final set returns nil.
func NewHealthCheckedRoundRobinKeyProvider(cfg KeyHealthConfig, keys ...string) (APIKeyProvider, error) {
	cleaned := cleanValues(keys...)
	if len(cleaned) == 0 {
		return nil, nil
	}

	resolved, err := resolveKeyHealthConfig(cfg)
	if err != nil {
		return nil, err
	}

	indexByKey := make(map[string]int, len(cleaned))
	for i, key := range cleaned {
		indexByKey[key] = i
	}

	return &healthCheckedRoundRobinKeyProvider{
		keys:          cleaned,
		states:        make([]keyHealthState, len(cleaned)),
		indexByKey:    indexByKey,
		failureThresh: resolved.FailureThreshold,
		cooldown:      resolved.Cooldown,
	}, nil
}

// NewStaticAccessTokenProvider creates a provider that always returns the same access token.
func NewStaticAccessTokenProvider(token string) AccessTokenProvider {
	provider := newRoundRobinValuesProvider(token)
	if provider == nil {
		return nil
	}
	return provider
}

// NewRoundRobinAccessTokenProvider creates a provider that rotates across access tokens.
// Empty values are ignored. If no tokens remain, nil is returned.
func NewRoundRobinAccessTokenProvider(tokens ...string) AccessTokenProvider {
	return newRoundRobinValuesProvider(tokens...)
}

// InjectCredentials clones query values and sets the Steam request credentials.
func InjectCredentials(query url.Values, apiKey, accessToken string) url.Values {
	cloned := make(url.Values, len(query)+2)
	for key, values := range query {
		copied := make([]string, len(values))
		copy(copied, values)
		cloned[key] = copied
	}
	if apiKey != "" && strings.TrimSpace(cloned.Get("key")) == "" {
		cloned.Set("key", apiKey)
	}
	if accessToken != "" && strings.TrimSpace(cloned.Get("access_token")) == "" {
		cloned.Set("access_token", accessToken)
	}
	return cloned
}

type roundRobinValuesProvider struct {
	values []string
	next   atomic.Uint64
}

func (p *roundRobinValuesProvider) Next(*http.Request) (string, error) {
	current := p.next.Add(1) - 1
	return p.values[current%uint64(len(p.values))], nil
}

func newRoundRobinValuesProvider(values ...string) *roundRobinValuesProvider {
	cleaned := cleanValues(values...)
	if len(cleaned) == 0 {
		return nil
	}
	return &roundRobinValuesProvider{values: cleaned}
}

type keyHealthState struct {
	failureScore  int
	cooldownUntil time.Time
}

type healthCheckedRoundRobinKeyProvider struct {
	mu            sync.Mutex
	keys          []string
	states        []keyHealthState
	indexByKey    map[string]int
	failureThresh int
	cooldown      time.Duration
	nextIndex     int
}

func (p *healthCheckedRoundRobinKeyProvider) Next(*http.Request) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	fallback := p.nextIndex % len(p.keys)
	for i := 0; i < len(p.keys); i++ {
		idx := (p.nextIndex + i) % len(p.keys)
		if state := p.states[idx]; !state.cooldownUntil.IsZero() && now.Before(state.cooldownUntil) {
			continue
		}
		p.nextIndex = (idx + 1) % len(p.keys)
		return p.keys[idx], nil
	}

	p.nextIndex = (fallback + 1) % len(p.keys)
	return p.keys[fallback], nil
}

func (p *healthCheckedRoundRobinKeyProvider) ReportAPIKeyResult(_ *http.Request, key string, statusCode int, err error) {
	if err != nil || key == "" {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	idx, ok := p.indexByKey[key]
	if !ok {
		return
	}

	state := &p.states[idx]
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusTooManyRequests {
		state.failureScore++
		if state.failureScore >= p.failureThresh {
			state.failureScore = 0
			state.cooldownUntil = time.Now().Add(p.cooldown)
		}
		return
	}

	state.failureScore = 0
	state.cooldownUntil = time.Time{}
}

func cleanValues(values ...string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}

func resolveKeyHealthConfig(cfg KeyHealthConfig) (KeyHealthConfig, error) {
	resolved := KeyHealthConfig{
		FailureThreshold: 2,
		Cooldown:         30 * time.Second,
	}

	switch {
	case cfg.FailureThreshold < 0:
		return KeyHealthConfig{}, fmt.Errorf("api key failure threshold must not be negative")
	case cfg.FailureThreshold > 0:
		resolved.FailureThreshold = cfg.FailureThreshold
	}

	switch {
	case cfg.Cooldown < 0:
		return KeyHealthConfig{}, fmt.Errorf("api key cooldown must not be negative")
	case cfg.Cooldown > 0:
		resolved.Cooldown = cfg.Cooldown
	}

	return resolved, nil
}
