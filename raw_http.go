package steam

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	itraffic "github.com/gofurry/steam-go/internal/traffic"
)

// RawHTTPRequestOptions controls raw addon-oriented HTTP execution through the SDK runtime.
type RawHTTPRequestOptions struct {
	TrafficClass         TrafficClass
	CookieJar            http.CookieJar
	MaxResponseBodyBytes int64
	// Retryable overrides the SDK retry default for this raw request.
	// nil means GET, HEAD, and OPTIONS are retryable; other methods are not.
	Retryable *bool
	// HostPolicy optionally restricts which absolute request hosts may be used.
	HostPolicy RawHTTPHostPolicy
}

// RawHTTPHostPolicy validates one raw HTTP request host before execution.
type RawHTTPHostPolicy interface {
	Allow(req *http.Request) error
}

type allowedRawHTTPHostPolicy struct {
	hosts map[string]struct{}
}

type suffixRawHTTPHostPolicy struct {
	suffixes []string
}

type anyRawHTTPHostPolicy struct {
	policies []RawHTTPHostPolicy
}

// RawHTTPBlockResult exposes block-detection metadata without discarding the raw response.
type RawHTTPBlockResult struct {
	Kind      ErrorKind
	Message   string
	Retryable bool
}

// RawHTTPResult exposes one raw HTTP response after retries, cache, and block detection.
type RawHTTPResult struct {
	StatusCode int
	Header     http.Header
	FinalURL   *url.URL
	Body       []byte
	Block      *RawHTTPBlockResult
}

// DoRawHTTPRequest executes one absolute HTTP request through one SDK traffic class runtime.
func (c *Client) DoRawHTTPRequest(ctx context.Context, req *http.Request, opts *RawHTTPRequestOptions) (RawHTTPResult, error) {
	if c == nil {
		return RawHTTPResult{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "client must not be nil", nil, nil)
	}
	if req == nil {
		return RawHTTPResult{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "request is required", nil, nil)
	}
	if req.URL == nil || req.URL.Scheme == "" || req.URL.Host == "" {
		return RawHTTPResult{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "request url must be absolute", nil, nil)
	}
	if opts != nil && opts.HostPolicy != nil {
		if err := opts.HostPolicy.Allow(req); err != nil {
			return RawHTTPResult{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "raw http host rejected", nil, err)
		}
	}

	class := TrafficClassOfficialAPI
	maxResponseBodyBytes := c.maxResponseBodyBytes
	if opts != nil {
		class = normalizeTrafficClass(opts.TrafficClass)
		if opts.MaxResponseBodyBytes < 0 {
			return RawHTTPResult{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max response body bytes must not be negative", nil, nil)
		}
		if opts.MaxResponseBodyBytes > 0 {
			maxResponseBodyBytes = opts.MaxResponseBodyBytes
		}
	}
	if maxResponseBodyBytes <= 0 {
		return RawHTTPResult{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max response body bytes must be greater than zero", nil, nil)
	}

	if ctx == nil {
		ctx = context.Background()
	}
	ctx = itraffic.WithClass(ctx, normalizeTrafficClass(class))
	if opts != nil && opts.CookieJar != nil {
		ctx = request.WithRuntimeCookieJar(ctx, opts.CookieJar)
	}

	var retryable *bool
	if opts != nil {
		retryable = opts.Retryable
	}
	result, err := request.ExecuteRawHTTPRequest(ctx, req, maxResponseBodyBytes, c.rawExecutionPolicy(normalizeTrafficClass(class)), retryable)
	if err != nil {
		return RawHTTPResult{}, err
	}
	return mapRawHTTPResult(result), nil
}

// NewAllowedRawHTTPHostPolicy returns a raw HTTP policy that allows only exact host matches.
// Hosts may be plain host names, host:port values, or absolute URLs.
func NewAllowedRawHTTPHostPolicy(hosts ...string) (RawHTTPHostPolicy, error) {
	allowed := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		normalized, err := normalizeRawHTTPPolicyHost(host)
		if err != nil {
			return nil, err
		}
		allowed[normalized] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, fmt.Errorf("at least one raw http host is required")
	}
	return allowedRawHTTPHostPolicy{hosts: allowed}, nil
}

// NewSteamRawHTTPHostPolicy returns a raw HTTP policy for common Steam API, Web, and static hosts.
func NewSteamRawHTTPHostPolicy() RawHTTPHostPolicy {
	policy, _ := NewAllowedRawHTTPHostPolicy(
		"api.steampowered.com",
		"store.steampowered.com",
		"steamcommunity.com",
		"shared.steamstatic.com",
		"cdn.cloudflare.steamstatic.com",
	)
	return policy
}

// NewSuffixRawHTTPHostPolicy returns a raw HTTP policy that allows exact host
// matches or subdomains under the configured suffixes.
func NewSuffixRawHTTPHostPolicy(suffixes ...string) (RawHTTPHostPolicy, error) {
	allowed := make([]string, 0, len(suffixes))
	seen := make(map[string]struct{}, len(suffixes))
	for _, suffix := range suffixes {
		normalized, err := normalizeRawHTTPPolicySuffix(suffix)
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
	return suffixRawHTTPHostPolicy{suffixes: allowed}, nil
}

// NewSteamStaticRawHTTPHostPolicy returns a raw HTTP policy for common Steam static/CDN hosts.
func NewSteamStaticRawHTTPHostPolicy() RawHTTPHostPolicy {
	suffixPolicy, _ := NewSuffixRawHTTPHostPolicy("steamstatic.com")
	exactPolicy, _ := NewAllowedRawHTTPHostPolicy("steamcdn-a.akamaihd.net")
	return anyRawHTTPHostPolicy{policies: []RawHTTPHostPolicy{suffixPolicy, exactPolicy}}
}

func (p allowedRawHTTPHostPolicy) Allow(req *http.Request) error {
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

func (p suffixRawHTTPHostPolicy) Allow(req *http.Request) error {
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

func (p anyRawHTTPHostPolicy) Allow(req *http.Request) error {
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

func normalizeRawHTTPPolicyHost(raw string) (string, error) {
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

func normalizeRawHTTPPolicySuffix(raw string) (string, error) {
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

func (c *Client) rawExecutionPolicy(class TrafficClass) request.ExecutionPolicy {
	class = normalizeTrafficClass(class)
	if policy, ok := c.rawRuntimes.classPolicies[itraffic.Class(class)]; ok {
		return policy
	}
	return c.rawRuntimes.defaultPolicy
}

func mapRawHTTPResult(src request.HTTPResult) RawHTTPResult {
	out := RawHTTPResult{
		StatusCode: src.StatusCode,
		Header:     cloneRawHTTPHeader(src.Header),
		FinalURL:   cloneRawHTTPURL(src.FinalURL),
		Body:       cloneRawHTTPBody(src.Body),
	}
	if src.Block != nil {
		out.Block = &RawHTTPBlockResult{
			Kind:      src.Block.ErrorKind,
			Message:   src.Block.Message,
			Retryable: src.Block.Retryable,
		}
	}
	return out
}

func cloneRawHTTPHeader(src http.Header) http.Header {
	if len(src) == 0 {
		return nil
	}
	cloned := make(http.Header, len(src))
	for key, values := range src {
		copied := make([]string, len(values))
		copy(copied, values)
		cloned[key] = copied
	}
	return cloned
}

func cloneRawHTTPURL(src *url.URL) *url.URL {
	if src == nil {
		return nil
	}
	cloned := *src
	return &cloned
}

func cloneRawHTTPBody(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	cloned := make([]byte, len(src))
	copy(cloned, src)
	return cloned
}
