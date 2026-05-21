package assets

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gofurry/steam-go/addons/assets/internal/httpasset"
)

// DownloadURLs downloads one or more URLs into dir.
//
// File names are taken from each URL path. Batch downloads try every URL; successful
// files remain on disk even when later downloads fail. Failed items are marked in
// DownloadResult.Error and returned together as one joined error.
func DownloadURLs(ctx context.Context, dir string, urls ...string) ([]DownloadResult, error) {
	return DownloadURLsWithOptions(ctx, DownloadOptions{Dir: dir}, urls...)
}

// DownloadURLsWithClient is DownloadURLs with a caller-supplied HTTP client.
func DownloadURLsWithClient(ctx context.Context, client *http.Client, dir string, urls ...string) ([]DownloadResult, error) {
	return DownloadURLsWithOptions(ctx, DownloadOptions{Dir: dir, HTTPClient: client}, urls...)
}

// DownloadURLsWithOptions downloads one or more URLs with explicit options.
func DownloadURLsWithOptions(ctx context.Context, opts DownloadOptions, urls ...string) ([]DownloadResult, error) {
	if opts.Dir == "" {
		return nil, fmt.Errorf("download dir must not be empty")
	}
	requests := make([]downloadRequest, 0, len(urls))
	for _, rawURL := range urls {
		name, err := httpasset.Filename(rawURL)
		if err != nil {
			requests = append(requests, downloadRequest{
				item: URLItem{URL: rawURL},
				url:  rawURL,
				err:  fmt.Errorf("download %s: %w", rawURL, err),
			})
			continue
		}
		requests = append(requests, downloadRequest{
			item: URLItem{URL: rawURL},
			url:  rawURL,
			path: filepath.Join(opts.Dir, name),
		})
	}
	return downloadRequests(ctx, opts.HTTPClient, effectiveOverwrite(opts.Overwrite, opts.SkipExisting), opts.Concurrency, requests)
}

// DownloadAppAssets builds and downloads app asset URLs into opts.Dir.
//
// If opts.Kinds is empty, all standard Store and Library asset kinds are downloaded.
// opts.Mode defaults to StoreFlat. StoreFlat prefixes generated filenames with the AppID;
// StoreByAppID writes files into a child directory named by AppID.
// Batch downloads try every generated URL; successful files remain on disk even when
// some assets are missing. Failed items are marked in DownloadResult.Error and returned
// together as one joined error.
func DownloadAppAssets(ctx context.Context, opts DownloadAppOptions, appIDs ...uint32) ([]DownloadResult, error) {
	if opts.Dir == "" {
		return nil, fmt.Errorf("download dir must not be empty")
	}
	mode := opts.Mode
	if mode == "" {
		mode = StoreFlat
	}
	if mode != StoreFlat && mode != StoreByAppID {
		return nil, fmt.Errorf("unknown store mode %q", mode)
	}
	if err := validateDownloadLanguage(opts); err != nil {
		return nil, err
	}

	requests := appDownloadRequests(opts, mode, appIDs)
	return downloadRequests(ctx, opts.HTTPClient, effectiveOverwrite(opts.Overwrite, opts.SkipExisting), opts.Concurrency, requests)
}

type downloadRequest struct {
	item URLItem
	url  string
	path string
	err  error
}

func appDownloadRequests(opts DownloadAppOptions, mode StoreMode, appIDs []uint32) []downloadRequest {
	kinds := opts.Kinds
	if len(kinds) == 0 {
		kinds = allDownloadKinds(opts.Language)
	}

	items := ListKindsWithLanguage(opts.Language, kinds, appIDs...)
	out := make([]downloadRequest, 0, len(items))
	for _, item := range items {
		name := appAssetFilename(item.AppID, item.Kind, item.URL, opts.Language, opts.FilenameStyle)
		out = append(out, downloadRequest{
			item: item,
			url:  item.URL,
			path: appDownloadPath(opts.Dir, mode, item.AppID, name),
		})
	}
	return out
}

func allDownloadKinds(language string) []Kind {
	kinds := DefaultKinds()
	if language != "" {
		kinds = append([]Kind{KindHeaderLocalized}, kinds...)
	}
	return kinds
}

func validateDownloadLanguage(opts DownloadAppOptions) error {
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

func containsKind(kinds []Kind, target Kind) bool {
	for _, kind := range kinds {
		if kind == target {
			return true
		}
	}
	return false
}

func appDownloadPath(dir string, mode StoreMode, appID uint32, name string) string {
	appIDText := strconv.FormatUint(uint64(appID), 10)
	if mode == StoreByAppID {
		return filepath.Join(dir, appIDText, name)
	}
	return filepath.Join(dir, appIDText+"_"+name)
}

func downloadRequests(ctx context.Context, client *http.Client, overwrite OverwriteMode, concurrency int, requests []downloadRequest) ([]DownloadResult, error) {
	if concurrency <= 0 {
		concurrency = 1
	}
	results := make([]DownloadResult, len(requests))
	errs := make([]error, len(requests))

	jobs := make(chan int)
	var wg sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				result, err := downloadRequestOne(ctx, client, overwrite, requests[index])
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

func downloadRequestOne(ctx context.Context, client *http.Client, overwrite OverwriteMode, req downloadRequest) (DownloadResult, error) {
	if req.err != nil {
		return DownloadResult{
			AppID:  req.item.AppID,
			Kind:   req.item.Kind,
			ID:     req.item.ID,
			Name:   req.item.Name,
			URL:    req.url,
			Path:   req.path,
			Status: DownloadStatusFailed,
			Error:  req.err.Error(),
		}, req.err
	}
	if skipped, result, err := maybeSkipExisting(req, overwrite); skipped || err != nil {
		return result, err
	}

	result, err := httpasset.Download(ctx, client, req.url, req.path)
	status := DownloadStatusDownloaded
	if err != nil {
		status = DownloadStatusFailed
	}
	return DownloadResult{
		AppID:         req.item.AppID,
		Kind:          req.item.Kind,
		ID:            req.item.ID,
		Name:          req.item.Name,
		URL:           result.URL,
		Path:          req.path,
		Status:        status,
		StatusCode:    result.StatusCode,
		ContentType:   result.ContentType,
		ContentLength: result.ContentLength,
		BytesWritten:  result.BytesWritten,
		Error:         errorString(err),
	}, err
}

func maybeSkipExisting(req downloadRequest, overwrite OverwriteMode) (bool, DownloadResult, error) {
	info, err := os.Stat(req.path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, DownloadResult{}, nil
		}
		result := failedDownloadResult(req, err)
		return true, result, err
	}
	if info.IsDir() {
		err := fmt.Errorf("destination is a directory")
		return true, failedDownloadResult(req, err), err
	}
	if overwrite == OverwriteNever || (overwrite == OverwriteIfZero && info.Size() > 0) {
		return true, DownloadResult{
			AppID:         req.item.AppID,
			Kind:          req.item.Kind,
			ID:            req.item.ID,
			Name:          req.item.Name,
			URL:           req.url,
			Path:          req.path,
			Status:        DownloadStatusSkipped,
			ContentLength: info.Size(),
		}, nil
	}
	return false, DownloadResult{}, nil
}

func failedDownloadResult(req downloadRequest, err error) DownloadResult {
	return DownloadResult{
		AppID:  req.item.AppID,
		Kind:   req.item.Kind,
		ID:     req.item.ID,
		Name:   req.item.Name,
		URL:    req.url,
		Path:   req.path,
		Status: DownloadStatusFailed,
		Error:  err.Error(),
	}
}

func effectiveOverwrite(mode OverwriteMode, skipExisting bool) OverwriteMode {
	if mode != "" {
		return mode
	}
	if skipExisting {
		return OverwriteNever
	}
	return OverwriteAlways
}

func appAssetFilename(appID uint32, kind Kind, rawURL, language string, style FilenameStyle) string {
	if style == "" {
		style = FilenameOriginal
	}
	original, err := httpasset.Filename(rawURL)
	if err != nil {
		original = kindFilename(kind, language)
	}
	switch style {
	case FilenameKind:
		return kindFilename(kind, language)
	case FilenameAppKind:
		return strconv.FormatUint(uint64(appID), 10) + "_" + kindFilename(kind, language)
	default:
		return original
	}
}

func kindFilename(kind Kind, language string) string {
	switch kind {
	case KindHeader:
		return "header.jpg"
	case KindHeaderLocalized:
		return "header_" + language + ".jpg"
	case KindCapsuleSmall:
		return "capsule_small.jpg"
	case KindCapsuleMain:
		return "capsule_main.jpg"
	case KindLibraryCapsule:
		return "library_capsule.jpg"
	case KindLibraryCapsule2x:
		return "library_capsule_2x.jpg"
	case KindLibraryHero:
		return "library_hero.jpg"
	case KindLibraryLogo:
		return "library_logo.png"
	case KindLibraryLogo2x:
		return "library_logo_2x.png"
	default:
		return string(kind)
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
