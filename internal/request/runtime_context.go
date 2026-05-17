package request

import (
	"context"
	"net/http"

	"github.com/gofurry/steam-go/internal/traffic"
)

// WithRuntimeCookieJar attaches one runtime cookie jar override to a context.
func WithRuntimeCookieJar(ctx context.Context, jar http.CookieJar) context.Context {
	return traffic.WithCookieJar(ctx, jar)
}

// RuntimeCookieJarFromContext resolves one runtime cookie jar override from context.
func RuntimeCookieJarFromContext(ctx context.Context) (http.CookieJar, bool) {
	return traffic.CookieJarFromContext(ctx)
}
