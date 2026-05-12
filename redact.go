package steam

import "net/url"

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
	query.Del("key")
	query.Del("access_token")
	cloned.RawQuery = query.Encode()
	return &cloned
}
