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
// use ReadURLsWithOptions to change it. For large batches, prefer ReadEachURLs
// so data can be processed and released one item at a time.
//
// When URLs come from untrusted input, use ReadURLsWithOptions with a
// URLValidator.
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
		err := validateDirectURL(rawURL, opts.URLValidator)
		requests = append(requests, readRequest{
			item: URLItem{URL: rawURL},
			url:  rawURL,
			err:  err,
		})
	}
	return readRequests(ctx, opts.HTTPClient, effectiveReadMaxBytes(opts.MaxBytes), opts.Concurrency, requests)
}

// ReadEachURLs reads direct URLs and calls handle for each result without
// retaining the full batch in memory.
//
// The handler is called for successful and failed items. When Concurrency is
// greater than 1, reads may finish out of order; handler calls are serialized.
func ReadEachURLs(ctx context.Context, opts ReadOptions, handle ReadHandler, urls ...string) error {
	requests := make([]readRequest, 0, len(urls))
	for _, rawURL := range urls {
		err := validateDirectURL(rawURL, opts.URLValidator)
		requests = append(requests, readRequest{
			item: URLItem{URL: rawURL},
			url:  rawURL,
			err:  err,
		})
	}
	return readEachRequests(ctx, opts.HTTPClient, effectiveReadMaxBytes(opts.MaxBytes), opts.Concurrency, requests, handle)
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

// ReadEachAppAssets builds app asset URLs and calls handle for each result
// without retaining the full batch in memory.
func ReadEachAppAssets(ctx context.Context, opts ReadAppOptions, handle ReadHandler, appIDs ...uint32) error {
	if err := validateReadLanguage(opts); err != nil {
		return err
	}
	kinds := opts.Kinds
	if len(kinds) == 0 {
		kinds = allDownloadKinds(opts.Language)
	}
	items := ListKindsWithLanguage(opts.Language, kinds, appIDs...)
	return readEachURLItems(ctx, opts.HTTPClient, opts.MaxBytes, opts.Concurrency, items, handle)
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

// ReadEachStoreMedia fetches Storefront media URLs and calls handle for each
// result without retaining the full batch in memory.
func ReadEachStoreMedia(ctx context.Context, service *storefront.Service, opts ReadStoreMediaOptions, handle ReadHandler, appIDs ...uint32) error {
	items, err := FetchStoreMediaURLs(ctx, service, StoreMediaOptions{
		CountryCode: opts.CountryCode,
		Language:    opts.Language,
		Kinds:       opts.Kinds,
	}, appIDs...)
	if err != nil {
		return err
	}
	return readEachURLItems(ctx, opts.HTTPClient, opts.MaxBytes, opts.Concurrency, items, handle)
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

func readEachURLItems(ctx context.Context, client *http.Client, maxBytes int64, concurrency int, items []URLItem, handle ReadHandler) error {
	requests := make([]readRequest, 0, len(items))
	for _, item := range items {
		requests = append(requests, readRequest{
			item: item,
			url:  item.URL,
		})
	}
	return readEachRequests(ctx, client, effectiveReadMaxBytes(maxBytes), concurrency, requests, handle)
}

type readRequest struct {
	item URLItem
	url  string
	err  error
}

func readRequests(ctx context.Context, client *http.Client, maxBytes int64, concurrency int, requests []readRequest) ([]ReadResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if concurrency <= 0 {
		concurrency = 1
	}
	results := make([]ReadResult, len(requests))
	errs := make([]error, len(requests))
	submitted := make([]bool, len(requests))

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
enqueue:
	for index := range requests {
		select {
		case <-ctx.Done():
			break enqueue
		case jobs <- index:
			submitted[index] = true
		}
	}
	close(jobs)
	wg.Wait()

	if err := ctx.Err(); err != nil {
		for index, ok := range submitted {
			if ok {
				continue
			}
			results[index] = canceledReadResult(requests[index], err)
			errs[index] = err
		}
	}

	joined := make([]error, 0)
	for _, err := range errs {
		if err != nil {
			joined = append(joined, err)
		}
	}
	return results, errors.Join(joined...)
}

func canceledReadResult(req readRequest, err error) ReadResult {
	return ReadResult{
		AppID:    req.item.AppID,
		Kind:     req.item.Kind,
		ID:       req.item.ID,
		Name:     req.item.Name,
		URL:      req.url,
		Digest:   req.item.Digest,
		Filename: req.item.Filename,
		Source:   req.item.Source,
		Error:    err.Error(),
	}
}

func readRequestOne(ctx context.Context, client *http.Client, maxBytes int64, req readRequest) (ReadResult, error) {
	if req.err != nil {
		return ReadResult{
			AppID:    req.item.AppID,
			Kind:     req.item.Kind,
			ID:       req.item.ID,
			Name:     req.item.Name,
			URL:      req.url,
			Digest:   req.item.Digest,
			Filename: req.item.Filename,
			Source:   req.item.Source,
			Error:    req.err.Error(),
		}, req.err
	}
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
		Digest:        req.item.Digest,
		Filename:      req.item.Filename,
		Source:        req.item.Source,
		StatusCode:    result.StatusCode,
		ContentType:   result.ContentType,
		ContentLength: result.ContentLength,
		BytesRead:     int64(len(data)),
		Data:          data,
		Error:         errorString(err),
	}, err
}

func readEachRequests(ctx context.Context, client *http.Client, maxBytes int64, concurrency int, requests []readRequest, handle ReadHandler) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if handle == nil {
		return fmt.Errorf("read handler must not be nil")
	}
	if concurrency <= 0 {
		concurrency = 1
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	var handleMu sync.Mutex
	var errMu sync.Mutex
	errs := make([]error, 0)

	appendErr := func(err error) {
		if err == nil {
			return
		}
		errMu.Lock()
		errs = append(errs, err)
		errMu.Unlock()
	}

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				result, err := readRequestOne(ctx, client, maxBytes, requests[index])
				appendErr(err)

				handleMu.Lock()
				handleErr := handle(result)
				handleMu.Unlock()
				if handleErr != nil {
					appendErr(fmt.Errorf("read handler for %s: %w", result.URL, handleErr))
				}
			}
		}()
	}
enqueue:
	for index := range requests {
		select {
		case <-ctx.Done():
			appendErr(ctx.Err())
			break enqueue
		case jobs <- index:
		}
	}
	close(jobs)
	wg.Wait()

	return errors.Join(errs...)
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
