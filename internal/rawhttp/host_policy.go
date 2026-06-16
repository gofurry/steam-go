package rawhttp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// HostPolicy validates one raw HTTP request host before execution.
type HostPolicy interface {
	Allow(req *http.Request) error
}

type allowedHostPolicy struct {
	hosts map[string]struct{}
}

type suffixHostPolicy struct {
	suffixes []string
}

type anyHostPolicy struct {
	policies []HostPolicy
}

// NewAllowedHostPolicy returns a policy that allows only exact host matches.
func NewAllowedHostPolicy(hosts ...string) (HostPolicy, error) {
	allowed := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		normalized, err := normalizeHost(host)
		if err != nil {
			return nil, err
		}
		allowed[normalized] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, fmt.Errorf("at least one raw http host is required")
	}
	return allowedHostPolicy{hosts: allowed}, nil
}

// NewSuffixHostPolicy returns a policy that allows exact host matches or subdomains under suffixes.
func NewSuffixHostPolicy(suffixes ...string) (HostPolicy, error) {
	allowed := make([]string, 0, len(suffixes))
	seen := make(map[string]struct{}, len(suffixes))
	for _, suffix := range suffixes {
		normalized, err := normalizeSuffix(suffix)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		allowed = append(allowed, normalized)
	}
	if len(allowed) == 0 {
		return nil, fmt.Errorf("at least one raw http host suffix is required")
	}
	return suffixHostPolicy{suffixes: allowed}, nil
}

// NewAnyHostPolicy returns a policy that allows a request when any child policy allows it.
func NewAnyHostPolicy(policies ...HostPolicy) HostPolicy {
	return anyHostPolicy{policies: policies}
}

func (p allowedHostPolicy) Allow(req *http.Request) error {
	if req == nil || req.URL == nil {
		return fmt.Errorf("request url is required")
	}
	host := strings.ToLower(strings.TrimSpace(req.URL.Host))
	if host == "" {
		return fmt.Errorf("request host is required")
	}
	if _, ok := p.hosts[host]; ok {
		return nil
	}
	return fmt.Errorf("host %q is not allowed", host)
}

func (p suffixHostPolicy) Allow(req *http.Request) error {
	if req == nil || req.URL == nil {
		return fmt.Errorf("request url is required")
	}
	host := strings.ToLower(strings.Trim(strings.TrimSpace(req.URL.Hostname()), "."))
	if host == "" {
		return fmt.Errorf("request host is required")
	}
	for _, suffix := range p.suffixes {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return nil
		}
	}
	return fmt.Errorf("host %q does not match allowed suffixes", host)
}

func (p anyHostPolicy) Allow(req *http.Request) error {
	if len(p.policies) == 0 {
		return fmt.Errorf("no raw http host policies configured")
	}
	var lastErr error
	for _, policy := range p.policies {
		if policy == nil {
			continue
		}
		if err := policy.Allow(req); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("no raw http host policies configured")
}

func normalizeHost(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("raw http host must not be empty")
	}
	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", fmt.Errorf("invalid raw http host %q: %w", raw, err)
		}
		raw = parsed.Host
	} else {
		if strings.ContainsAny(raw, "/?#") {
			return "", fmt.Errorf("raw http host %q must be a host or absolute url", raw)
		}
		if parsed, err := url.Parse("//" + raw); err == nil {
			raw = parsed.Host
		}
	}
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" || raw == "*" {
		return "", fmt.Errorf("raw http host must be a concrete host")
	}
	return raw, nil
}

func normalizeSuffix(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("raw http host suffix must not be empty")
	}

	var host string
	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", fmt.Errorf("invalid raw http host suffix %q: %w", raw, err)
		}
		if parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
			return "", fmt.Errorf("raw http host suffix %q must not include path, query, fragment, or userinfo", raw)
		}
		if parsed.Port() != "" {
			return "", fmt.Errorf("raw http host suffix %q must not include a port", raw)
		}
		host = parsed.Hostname()
	} else {
		if strings.ContainsAny(raw, "/?#@") {
			return "", fmt.Errorf("raw http host suffix %q must be a host suffix", raw)
		}
		parsed, err := url.Parse("//" + raw)
		if err != nil {
			return "", fmt.Errorf("invalid raw http host suffix %q: %w", raw, err)
		}
		if parsed.Port() != "" {
			return "", fmt.Errorf("raw http host suffix %q must not include a port", raw)
		}
		host = parsed.Hostname()
	}

	host = strings.ToLower(strings.Trim(strings.TrimSpace(host), "."))
	if host == "" || host == "*" || strings.Contains(host, "*") {
		return "", fmt.Errorf("raw http host suffix must be a concrete host suffix")
	}
	if strings.Contains(host, "..") {
		return "", fmt.Errorf("raw http host suffix %q is malformed", raw)
	}
	return host, nil
}
