package steam

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const redactedSensitiveValue = "[REDACTED]"

var sensitiveURLQueryKeys = map[string]struct{}{
	"access_token":         {},
	"api_key":              {},
	"apikey":               {},
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

var sensitiveURLFallbackPatterns = buildSensitiveURLFallbackPatterns()

var (
	authorizationTextPattern = regexp.MustCompile(`(?i)(\bauthorization\s*:\s*(?:bearer|basic)\s+)([^\s,;]+)`)
	proxyUserinfoTextPattern = regexp.MustCompile(`(?i)\b([a-z][a-z0-9+.-]*://)([^/@\s]+@)`)
)

// RedactSensitiveURL removes userinfo plus API key-like query parameters from one URL string.
//
// When parsing fails, a best-effort fallback redacts obvious credential-bearing key/value pairs.
func RedactSensitiveURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return redactSensitiveURLFallback(rawURL)
	}
	return redactSensitiveURL(parsed).String()
}

// RedactSensitiveText redacts obvious credential-bearing fragments in free-form text.
//
// It is best-effort and intentionally conservative; it is meant for logs,
// diagnostics, examples, and reports, not for classifying every possible secret.
func RedactSensitiveText(text string) string {
	redacted := redactSensitiveURLFallback(text)
	redacted = authorizationTextPattern.ReplaceAllString(redacted, `${1}`+redactedSensitiveValue)
	redacted = proxyUserinfoTextPattern.ReplaceAllString(redacted, `${1}`+redactedSensitiveValue+"@")
	return redacted
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

func buildSensitiveURLFallbackPatterns() []*regexp.Regexp {
	keys := make([]*regexp.Regexp, 0, len(sensitiveURLQueryKeys))
	for key := range sensitiveURLQueryKeys {
		keys = append(keys, regexp.MustCompile(`(?i)(^|[?&;\s])(`+regexp.QuoteMeta(key)+`)(=)([^&;\s]*)`))
	}
	return keys
}

func redactSensitiveURLFallback(raw string) string {
	redacted := raw
	for _, pattern := range sensitiveURLFallbackPatterns {
		redacted = pattern.ReplaceAllString(redacted, `${1}${2}${3}`+redactedSensitiveValue)
	}
	return redacted
}
