package assets

import (
	"context"
	"net/http"

	"github.com/gofurry/steam-go/addons/assets/internal/httpasset"
)

// VerifyURLs checks whether each URL currently resolves to a 2xx HTTP response.
//
// It uses HEAD first and falls back to GET when the server returns 405 or 501.
// HTTP non-2xx responses are reported as Exists=false and are not returned as errors.
//
// When URLs come from untrusted input, use VerifyURLsWithOptions with a
// URLValidator.
func VerifyURLs(ctx context.Context, urls ...string) ([]VerifyResult, error) {
	return VerifyURLsWithClient(ctx, nil, urls...)
}

// VerifyURLsWithClient is VerifyURLs with a caller-supplied HTTP client.
func VerifyURLsWithClient(ctx context.Context, client *http.Client, urls ...string) ([]VerifyResult, error) {
	return VerifyURLsWithOptions(ctx, VerifyOptions{HTTPClient: client}, urls...)
}

// VerifyURLsWithOptions verifies one or more direct URLs with explicit options.
func VerifyURLsWithOptions(ctx context.Context, opts VerifyOptions, urls ...string) ([]VerifyResult, error) {
	out := make([]VerifyResult, 0, len(urls))
	for _, rawURL := range urls {
		if err := validateDirectURL(rawURL, opts.URLValidator); err != nil {
			return out, err
		}
		verified, err := verifyURLItem(ctx, opts.HTTPClient, URLItem{URL: rawURL})
		if err != nil {
			return out, err
		}
		out = append(out, verified)
	}
	return out, nil
}

// VerifyAppAssets builds and verifies app asset URLs.
//
// If opts.Kinds is empty, all standard Store and Library asset kinds are verified.
func VerifyAppAssets(ctx context.Context, opts VerifyAppOptions, appIDs ...uint32) ([]VerifyResult, error) {
	kinds := opts.Kinds
	if len(kinds) == 0 {
		kinds = allDownloadKinds(opts.Language)
	}
	items := ListKindsWithLanguage(opts.Language, kinds, appIDs...)
	out := make([]VerifyResult, 0, len(items))
	for _, item := range items {
		verified, err := verifyURLItem(ctx, opts.HTTPClient, item)
		if err != nil {
			return out, err
		}
		out = append(out, verified)
	}
	return out, nil
}

func verifyURLItem(ctx context.Context, client *http.Client, item URLItem) (VerifyResult, error) {
	result, exists, err := httpasset.Verify(ctx, client, item.URL)
	if err != nil {
		return VerifyResult{}, err
	}
	return VerifyResult{
		AppID:         item.AppID,
		Kind:          item.Kind,
		ID:            item.ID,
		Name:          item.Name,
		URL:           result.URL,
		Exists:        exists,
		StatusCode:    result.StatusCode,
		ContentType:   result.ContentType,
		ContentLength: result.ContentLength,
	}, nil
}
