package auth

import (
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
)

// APIKeyProvider resolves a key for an outgoing request.
type APIKeyProvider interface {
	Next(req *http.Request) (string, error)
}

// AccessTokenProvider resolves an access token for an outgoing request.
type AccessTokenProvider interface {
	Next(req *http.Request) (string, error)
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
	if apiKey != "" {
		cloned.Set("key", apiKey)
	}
	if accessToken != "" {
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
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	if len(cleaned) == 0 {
		return nil
	}
	return &roundRobinValuesProvider{values: cleaned}
}
