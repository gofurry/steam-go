package steam

import (
	"net/http"
	"net/url"
	"strings"
)

const redactedSensitiveValue = "[REDACTED]"

var sensitiveURLQueryKeys = map[string]struct{}{
	"access_token":         {},
	"key":                  {},
	"loyalty_webapi_token": {},
	"refresh_token":        {},
	"sessionid":            {},
	"steamloginsecure":     {},
	"webapi_token":         {},
}

var sensitiveHeaderNames = map[string]struct{}{
	"authorization":           {},
	"cookie":                  {},
	"proxy-authorization":     {},
	"set-cookie":              {},
	"x-api-key":               {},
	"x-steam-web-api-key":     {},
	"x-webapi-key":            {},
	"x-web-api-key":           {},
	"x-webapi-token":          {},
	"x-steam-access-token":    {},
	"x-steam-refresh-token":   {},
	"x-steam-session-token":   {},
	"x-steam-login-secure":    {},
	"x-steam-publisher-key":   {},
	"x-steam-publisher-token": {},
}

// RedactSensitiveURL removes userinfo plus API key-like query parameters from one URL string.
//
// When parsing fails, the original input is returned unchanged so callers can safely use it in logs.
func RedactSensitiveURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return redactSensitiveURL(parsed).String()
}

func redactSensitiveURL(parsed *url.URL) *url.URL {
	if parsed == nil {
		return nil
	}

	cloned := *parsed
	cloned.User = nil

	query := cloned.Query()
	for key := range query {
		if isSensitiveURLQueryKey(key) {
			query.Del(key)
		}
	}
	cloned.RawQuery = query.Encode()
	return &cloned
}

// RedactSensitiveHeaderValue redacts one HTTP header value when the header is credential-bearing.
func RedactSensitiveHeaderValue(name, value string) string {
	if isSensitiveHeaderName(name) {
		return redactedSensitiveValue
	}
	return value
}

// RedactSensitiveHeaders returns a cloned header map with credential-bearing values redacted.
func RedactSensitiveHeaders(header http.Header) http.Header {
	if header == nil {
		return nil
	}

	redacted := make(http.Header, len(header))
	for name, values := range header {
		copied := make([]string, len(values))
		for i, value := range values {
			copied[i] = RedactSensitiveHeaderValue(name, value)
		}
		redacted[name] = copied
	}
	return redacted
}

func isSensitiveURLQueryKey(key string) bool {
	_, ok := sensitiveURLQueryKeys[strings.ToLower(strings.TrimSpace(key))]
	return ok
}

func isSensitiveHeaderName(name string) bool {
	_, ok := sensitiveHeaderNames[strings.ToLower(strings.TrimSpace(name))]
	return ok
}
