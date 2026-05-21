package assets

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// URLValidator checks whether a direct URL is allowed before the package sends
// an HTTP request for VerifyURLs, ReadURLs, or DownloadURLs helpers.
type URLValidator func(*url.URL) error

// AllowHosts returns a URLValidator that accepts only the exact host names.
//
// Host matching is case-insensitive and ignores URL ports.
func AllowHosts(hosts ...string) URLValidator {
	allowed := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		host = normalizeHost(host)
		if host != "" {
			allowed[host] = struct{}{}
		}
	}
	return func(parsed *url.URL) error {
		host := strings.ToLower(parsed.Hostname())
		if _, ok := allowed[host]; ok {
			return nil
		}
		return fmt.Errorf("url host %q is not allowed", parsed.Hostname())
	}
}

// AllowHostSuffixes returns a URLValidator that accepts exact hosts or
// subdomains under any supplied suffix, for example "steamstatic.com".
//
// Host matching is case-insensitive and ignores URL ports.
func AllowHostSuffixes(suffixes ...string) URLValidator {
	allowed := make([]string, 0, len(suffixes))
	for _, suffix := range suffixes {
		suffix = strings.TrimPrefix(normalizeHost(suffix), ".")
		if suffix != "" {
			allowed = append(allowed, suffix)
		}
	}
	return func(parsed *url.URL) error {
		host := strings.ToLower(parsed.Hostname())
		for _, suffix := range allowed {
			if host == suffix || strings.HasSuffix(host, "."+suffix) {
				return nil
			}
		}
		return fmt.Errorf("url host %q is not allowed", parsed.Hostname())
	}
}

// SteamStaticURLValidator accepts Steam static asset hosts.
func SteamStaticURLValidator(parsed *url.URL) error {
	return steamStaticURLValidator(parsed)
}

var steamStaticURLValidator = AllowHostSuffixes("steamstatic.com")

func validateDirectURL(rawURL string, validator URLValidator) error {
	parsed, err := parseDirectURL(rawURL)
	if err != nil {
		return err
	}
	if validator != nil {
		return validator(parsed)
	}
	return nil
}

func parseDirectURL(rawURL string) (*url.URL, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("url must not be empty")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("url scheme must be http or https")
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("url host must not be empty")
	}
	return parsed, nil
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if split, _, err := net.SplitHostPort(host); err == nil {
		host = split
	}
	return strings.ToLower(strings.Trim(host, "[]"))
}
