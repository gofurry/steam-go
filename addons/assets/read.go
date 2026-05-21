package assets

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gofurry/steam-go/addons/assets/internal/httpasset"
	"github.com/gofurry/steam-go/web/storefront"
)

const defaultReadMaxBytes int64 = 32 << 20

// ReadURLs reads one or more URLs into memory.
//
// Batch reads try every URL. Failed items are marked in ReadResult.Error and
// returned together as one joined error. The default per-resource limit is 32 MiB;
// use ReadURLsWithOptions to change it.
func ReadURLs(ctx context.Context, urls ...string) ([]ReadResult, error) {
	return ReadURLsWithOptions(ctx, ReadOptions{}, urls...)
}

// ReadURLsWithClient is ReadURLs with a caller-supplied HTTP client.
func ReadURLsWithClient(ctx context.Context, client *http.Client, urls ...string) ([]ReadResult, error) {
	return ReadURLsWithOptions(ctx, ReadOptions{HTTPClient: client}, urls...)
}

// ReadURLsWithOptions reads one or more URLs into memory with explicit options.
func ReadURLsWithOptions(ctx context.Context, opts ReadOptions, urls ...string) ([]ReadResult, error) {
	requests := make([]readRequest, 0, len(urls))
	for _, rawURL := range urls {
		requests = append(requests, readRequest{
			item: URLItem{URL: rawURL},
			url:  rawURL,
		})
	}
	return readRequests(ctx, opts.HTTPClient, effectiveReadMaxBytes(opts.MaxBytes), opts.Concurrency, requests)
}

// ReadAppAssets builds app asset URLs and reads them into memory.
//
// If opts.Kinds is empty, all standard Store and Library asset kinds are read.
func ReadAppAssets(ctx context.Context, opts ReadAppOptions, appIDs ...uint32) ([]ReadResult, error) {
	if err := validateReadLanguage(opts); err != nil {
		return nil, err
	}
	kinds := opts.Kinds
	if len(kinds) == 0 {
		kinds = allDownloadKinds(opts.Language)
	}
	items := ListKindsWithLanguage(opts.Language, kinds, appIDs...)
	return readURLItems(ctx, opts.HTTPClient, opts.MaxBytes, opts.Concurrency, items)
}

// ReadStoreMedia fetches Storefront media URLs and reads them into memory.
//
// This is intended for screenshots, backgrounds, thumbnails, and playlist files.
// For large direct movie files, set MaxBytes deliberately or prefer DownloadStoreMedia.
func ReadStoreMedia(ctx context.Context, service *storefront.Service, opts ReadStoreMediaOptions, appIDs ...uint32) ([]ReadResult, error) {
	items, err := FetchStoreMediaURLs(ctx, service, StoreMediaOptions{
		CountryCode: opts.CountryCode,
		Language:    opts.Language,
		Kinds:       opts.Kinds,
	}, appIDs...)
	if err != nil {
		return nil, err
	}
	return readURLItems(ctx, opts.HTTPClient, opts.MaxBytes, opts.Concurrency, items)
}

func readURLItems(ctx context.Context, client *http.Client, maxBytes int64, concurrency int, items []URLItem) ([]ReadResult, error) {
	requests := make([]readRequest, 0, len(items))
	for _, item := range items {
		requests = append(requests, readRequest{
			item: item,
			url:  item.URL,
		})
	}
	return readRequests(ctx, client, effectiveReadMaxBytes(maxBytes), concurrency, requests)
}

type readRequest struct {
	item URLItem
	url  string
}

func readRequests(ctx context.Context, client *http.Client, maxBytes int64, concurrency int, requests []readRequest) ([]ReadResult, error) {
	if concurrency <= 0 {
		concurrency = 1
	}
	results := make([]ReadResult, len(requests))
	errs := make([]error, len(requests))

	jobs := make(chan int)
	var wg sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				result, err := readRequestOne(ctx, client, maxBytes, requests[index])
				results[index] = result
				errs[index] = err
			}
		}()
	}
	for index := range requests {
		jobs <- index
	}
	close(jobs)
	wg.Wait()

	joined := make([]error, 0)
	for _, err := range errs {
		if err != nil {
			joined = append(joined, err)
		}
	}
	return results, errors.Join(joined...)
}

func readRequestOne(ctx context.Context, client *http.Client, maxBytes int64, req readRequest) (ReadResult, error) {
	result, data, err := httpasset.Read(ctx, client, req.url, maxBytes)
	resultURL := result.URL
	if resultURL == "" {
		resultURL = req.url
	}
	return ReadResult{
		AppID:         req.item.AppID,
		Kind:          req.item.Kind,
		ID:            req.item.ID,
		Name:          req.item.Name,
		URL:           resultURL,
		StatusCode:    result.StatusCode,
		ContentType:   result.ContentType,
		ContentLength: result.ContentLength,
		BytesRead:     int64(len(data)),
		Data:          data,
		Error:         errorString(err),
	}, err
}

func validateReadLanguage(opts ReadAppOptions) error {
	if opts.Language == "" {
		return nil
	}
	if len(opts.Kinds) == 0 || containsKind(opts.Kinds, KindHeaderLocalized) {
		if HeaderLocalizedURLs(opts.Language, 1)[0] == "" {
			return fmt.Errorf("language must be a safe path token")
		}
	}
	return nil
}

func effectiveReadMaxBytes(maxBytes int64) int64 {
	if maxBytes > 0 {
		return maxBytes
	}
	return defaultReadMaxBytes
}
