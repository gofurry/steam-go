package steam

import (
	"context"
	"net/http"

	"github.com/gofurry/steam-go/internal/request"
	"github.com/gofurry/steam-go/internal/requestprofile"
)

const sdkDefaultUserAgent = requestprofile.SDKDefaultUserAgent

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

// DefaultPublicStoreHeaderProfileZH returns one stable zh-CN oriented store-page profile.
func DefaultPublicStoreHeaderProfileZH() HeaderProfile {
	return HeaderProfile{
		UserAgent:               "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
		Accept:                  "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		AcceptLanguage:          "zh-CN,zh;q=0.9,en;q=0.8",
		AcceptEncoding:          "gzip",
		UpgradeInsecureRequests: "1",
		SecFetchDest:            "document",
		SecFetchMode:            "navigate",
		SecFetchSite:            "none",
	}
}

// DefaultPublicStoreHeaderProfileEN returns one stable en-US oriented store-page profile.
func DefaultPublicStoreHeaderProfileEN() HeaderProfile {
	return HeaderProfile{
		UserAgent:               "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
		Accept:                  "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		AcceptLanguage:          "en-US,en;q=0.9",
		AcceptEncoding:          "gzip",
		UpgradeInsecureRequests: "1",
		SecFetchDest:            "document",
		SecFetchMode:            "navigate",
		SecFetchSite:            "none",
	}
}

// WithRefererSource attaches one explicit referer source URL to a request context.
func WithRefererSource(ctx context.Context, rawURL string) context.Context {
	return requestprofile.WithRefererSource(ctx, rawURL)
}

// NewStaticRefererSelector returns one fixed Referer selector.
func NewStaticRefererSelector(rawURL string) (RefererSelector, error) {
	return requestprofile.NewStaticRefererSelector(rawURL)
}

// NewRoutingRefererSelector routes Referer values by host/path.
func NewRoutingRefererSelector(routes ...RefererRoute) (RefererSelector, error) {
	mapped := make([]requestprofile.RefererRoute, 0, len(routes))
	for _, route := range routes {
		mapped = append(mapped, requestprofile.RefererRoute{
			Host:       route.Host,
			PathPrefix: route.PathPrefix,
			RefererURL: route.RefererURL,
		})
	}
	return requestprofile.NewRoutingRefererSelector(mapped...)
}

// NewContextRefererSelector prefers one referer source stored in context and falls back when absent.
func NewContextRefererSelector(fallback RefererSelector) RefererSelector {
	return requestprofile.NewContextRefererSelector(fallback)
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

func buildRequestPreparer(profile *HeaderProfile, selector RefererSelector) request.RequestPreparer {
	return requestprofile.BuildRequestPreparer(mapHeaderProfile(profile), selector)
}

func mapHeaderProfile(profile *HeaderProfile) *requestprofile.HeaderProfile {
	cloned := cloneHeaderProfile(profile)
	if cloned == nil {
		return nil
	}
	return &requestprofile.HeaderProfile{
		UserAgent:               cloned.UserAgent,
		Accept:                  cloned.Accept,
		AcceptLanguage:          cloned.AcceptLanguage,
		AcceptEncoding:          cloned.AcceptEncoding,
		UpgradeInsecureRequests: cloned.UpgradeInsecureRequests,
		SecFetchDest:            cloned.SecFetchDest,
		SecFetchMode:            cloned.SecFetchMode,
		SecFetchSite:            cloned.SecFetchSite,
		Extra:                   cloned.Extra,
	}
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
