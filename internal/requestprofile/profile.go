package requestprofile

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const SDKDefaultUserAgent = "steam-go/1"

type refererSourceContextKey struct{}

type refererSourceContextValue struct {
	value string
	err   error
}

// HeaderProfile describes one lightweight browser-like default request profile.
type HeaderProfile struct {
	UserAgent               string
	Accept                  string
	AcceptLanguage          string
	AcceptEncoding          string
	UpgradeInsecureRequests string
	SecFetchDest            string
	SecFetchMode            string
	SecFetchSite            string
	Extra                   http.Header
}

// RefererSelector chooses one request Referer value.
type RefererSelector interface {
	Next(req *http.Request) (string, error)
}

// RefererRoute defines one host/path based Referer rule.
type RefererRoute struct {
	Host       string
	PathPrefix string
	RefererURL string
}

// WithRefererSource attaches one explicit referer source URL to a request context.
func WithRefererSource(ctx context.Context, rawURL string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ctx
	}

	normalized, err := normalizeRefererURL(rawURL)
	return context.WithValue(ctx, refererSourceContextKey{}, refererSourceContextValue{
		value: normalized,
		err:   err,
	})
}

// NewStaticRefererSelector returns one fixed Referer selector.
func NewStaticRefererSelector(rawURL string) (RefererSelector, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, nil
	}
	normalized, err := normalizeRefererURL(rawURL)
	if err != nil {
		return nil, err
	}
	return staticRefererSelector{value: normalized}, nil
}

// NewRoutingRefererSelector routes Referer values by host/path.
func NewRoutingRefererSelector(routes ...RefererRoute) (RefererSelector, error) {
	compiled := make([]compiledRefererRoute, 0, len(routes))
	for _, route := range routes {
		normalized := ""
		if strings.TrimSpace(route.RefererURL) != "" {
			var err error
			normalized, err = normalizeRefererURL(route.RefererURL)
			if err != nil {
				return nil, err
			}
		}

		compiled = append(compiled, compiledRefererRoute{
			host:       strings.ToLower(strings.TrimSpace(route.Host)),
			pathPrefix: strings.TrimSpace(route.PathPrefix),
			refererURL: normalized,
		})
	}
	if len(compiled) == 0 {
		return nil, nil
	}
	return routingRefererSelector{routes: compiled}, nil
}

// NewContextRefererSelector prefers one referer source stored in context and falls back when absent.
func NewContextRefererSelector(fallback RefererSelector) RefererSelector {
	return contextRefererSelector{fallback: fallback}
}

// BuildRequestPreparer builds a request mutation hook for header and Referer defaults.
func BuildRequestPreparer(profile *HeaderProfile, selector RefererSelector) func(req *http.Request) error {
	if profile == nil && selector == nil {
		return nil
	}
	clonedProfile := cloneHeaderProfile(profile)
	return func(req *http.Request) error {
		if req == nil {
			return nil
		}
		applyHeaderProfile(req, clonedProfile)
		return applyRefererSelector(req, selector)
	}
}

func cloneHeaderProfile(profile *HeaderProfile) *HeaderProfile {
	if profile == nil {
		return nil
	}
	cloned := *profile
	if profile.Extra != nil {
		cloned.Extra = cloneHTTPHeader(profile.Extra)
	}
	return &cloned
}

func applyHeaderProfile(req *http.Request, profile *HeaderProfile) {
	if req == nil || profile == nil {
		return
	}

	setHeaderIfMissing(req.Header, "Accept", profile.Accept)
	setHeaderIfMissing(req.Header, "Accept-Language", profile.AcceptLanguage)
	setHeaderIfMissing(req.Header, "Accept-Encoding", profile.AcceptEncoding)
	setHeaderIfMissing(req.Header, "Upgrade-Insecure-Requests", profile.UpgradeInsecureRequests)
	setHeaderIfMissing(req.Header, "Sec-Fetch-Dest", profile.SecFetchDest)
	setHeaderIfMissing(req.Header, "Sec-Fetch-Mode", profile.SecFetchMode)
	setHeaderIfMissing(req.Header, "Sec-Fetch-Site", profile.SecFetchSite)

	if profile.UserAgent != "" {
		current := req.Header.Get("User-Agent")
		if current == "" || current == SDKDefaultUserAgent {
			req.Header.Set("User-Agent", profile.UserAgent)
		}
	}

	for key, values := range profile.Extra {
		if req.Header.Get(key) != "" || len(values) == 0 {
			continue
		}
		copied := make([]string, len(values))
		copy(copied, values)
		req.Header[key] = copied
	}
}

func applyRefererSelector(req *http.Request, selector RefererSelector) error {
	if req == nil || selector == nil || req.Header.Get("Referer") != "" {
		return nil
	}
	referer, err := selector.Next(req)
	if err != nil {
		return err
	}
	referer = strings.TrimSpace(referer)
	if referer == "" {
		return nil
	}
	normalized, err := normalizeRefererURL(referer)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", normalized)
	return nil
}

func setHeaderIfMissing(header http.Header, key, value string) {
	if header == nil || value == "" || header.Get(key) != "" {
		return
	}
	header.Set(key, value)
}

func cloneHTTPHeader(header http.Header) http.Header {
	if header == nil {
		return nil
	}
	cloned := make(http.Header, len(header))
	for key, values := range header {
		copied := make([]string, len(values))
		copy(copied, values)
		cloned[key] = copied
	}
	return cloned
}

func normalizeRefererURL(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", fmt.Errorf("parse referer url %q: %w", rawURL, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("referer url must include scheme and host")
	}
	return parsed.String(), nil
}

func refererSourceFromContext(ctx context.Context) (string, error, bool) {
	if ctx == nil {
		return "", nil, false
	}
	value, ok := ctx.Value(refererSourceContextKey{}).(refererSourceContextValue)
	if !ok {
		return "", nil, false
	}
	if value.err != nil {
		return "", value.err, true
	}
	if value.value == "" {
		return "", nil, false
	}
	return value.value, nil, true
}

type staticRefererSelector struct {
	value string
}

func (s staticRefererSelector) Next(*http.Request) (string, error) {
	return s.value, nil
}

type compiledRefererRoute struct {
	host       string
	pathPrefix string
	refererURL string
}

type routingRefererSelector struct {
	routes []compiledRefererRoute
}

func (s routingRefererSelector) Next(req *http.Request) (string, error) {
	if req == nil || req.URL == nil {
		return "", nil
	}
	host := strings.ToLower(req.URL.Hostname())
	path := req.URL.Path
	for _, route := range s.routes {
		if route.host != "" && route.host != host {
			continue
		}
		if route.pathPrefix != "" && !strings.HasPrefix(path, route.pathPrefix) {
			continue
		}
		return route.refererURL, nil
	}
	return "", nil
}

type contextRefererSelector struct {
	fallback RefererSelector
}

func (s contextRefererSelector) Next(req *http.Request) (string, error) {
	if req != nil {
		if value, err, ok := refererSourceFromContext(req.Context()); ok {
			return value, err
		}
	}
	if s.fallback == nil {
		return "", nil
	}
	return s.fallback.Next(req)
}
