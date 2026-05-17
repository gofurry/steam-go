package traffic

import (
	"context"
	"net/http"
)

// Class identifies one request traffic category inside the SDK.
type Class string

const (
	ClassOfficialAPI     Class = "official_api"
	ClassPublicStorePage Class = "public_store_page"
	ClassCommunityWeb    Class = "community_web"
	ClassMarketWeb       Class = "market_web"
)

type classContextKey struct{}
type requestSessionContextKey struct{}
type blockDetectionContextKey struct{}
type cookieJarContextKey struct{}

// WithClass attaches one traffic class to a context.
func WithClass(ctx context.Context, class Class) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, classContextKey{}, NormalizeClass(class))
}

// ClassFromContext resolves one traffic class from context.
func ClassFromContext(ctx context.Context) (Class, bool) {
	if ctx == nil {
		return "", false
	}
	class, ok := ctx.Value(classContextKey{}).(Class)
	if !ok {
		return "", false
	}
	return NormalizeClass(class), true
}

// NormalizeClass coerces empty or unknown values back to the default official API class.
func NormalizeClass(class Class) Class {
	switch class {
	case ClassPublicStorePage:
		return ClassPublicStorePage
	case ClassCommunityWeb:
		return ClassCommunityWeb
	case ClassMarketWeb:
		return ClassMarketWeb
	case ClassOfficialAPI:
		fallthrough
	default:
		return ClassOfficialAPI
	}
}

// WithRequestSessionKey attaches one explicit request-session key to a context.
func WithRequestSessionKey(ctx context.Context, key string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, requestSessionContextKey{}, key)
}

// RequestSessionKeyFromContext resolves one request-session key from context.
func RequestSessionKeyFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	key, ok := ctx.Value(requestSessionContextKey{}).(string)
	if !ok || key == "" {
		return "", false
	}
	return key, true
}

// WithBlockDetection attaches one block-detection enabled marker to a context.
func WithBlockDetection(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, blockDetectionContextKey{}, true)
}

// BlockDetectionFromContext resolves whether block detection is enabled for this request.
func BlockDetectionFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	enabled, _ := ctx.Value(blockDetectionContextKey{}).(bool)
	return enabled
}

// WithCookieJar attaches one runtime cookie jar override to a context.
func WithCookieJar(ctx context.Context, jar http.CookieJar) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if jar == nil {
		return ctx
	}
	return context.WithValue(ctx, cookieJarContextKey{}, jar)
}

// CookieJarFromContext resolves one runtime cookie jar override from context.
func CookieJarFromContext(ctx context.Context) (http.CookieJar, bool) {
	if ctx == nil {
		return nil, false
	}
	jar, ok := ctx.Value(cookieJarContextKey{}).(http.CookieJar)
	return jar, ok && jar != nil
}
