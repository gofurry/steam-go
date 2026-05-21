package assets

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/gofurry/steam-go/addons/assets/internal/httpasset"
	"github.com/gofurry/steam-go/web/storefront"
)

// FetchStoreMediaURLs requests Storefront appdetails and returns screenshot,
// movie, and background URLs. Results follow the requested kind order; repeated
// screenshots and movies keep Steam's response order within each kind.
//
// If opts.Kinds is empty, StoreMediaKinds is used.
func FetchStoreMediaURLs(ctx context.Context, service *storefront.Service, opts StoreMediaOptions, appIDs ...uint32) ([]URLItem, error) {
	if service == nil {
		return nil, fmt.Errorf("storefront service must not be nil")
	}
	kinds := opts.Kinds
	if len(kinds) == 0 {
		kinds = StoreMediaKinds()
	}

	out := make([]URLItem, 0)
	for _, appID := range appIDs {
		if appID == 0 {
			return out, fmt.Errorf("app id must be greater than zero")
		}
		envelope, err := service.GetAppDetails(ctx, appID, &storefront.GetAppDetailsOptions{
			CountryCode: opts.CountryCode,
			Language:    opts.Language,
		})
		if err != nil {
			return out, err
		}
		result, ok := envelope[strconv.FormatUint(uint64(appID), 10)]
		if !ok {
			return out, fmt.Errorf("appdetails response did not include app id %d", appID)
		}
		if !result.Success {
			return out, fmt.Errorf("appdetails lookup for app id %d was not successful", appID)
		}
		out = append(out, storeMediaItems(appID, result.Data, kinds)...)
	}
	return out, nil
}

// VerifyStoreMedia fetches Storefront media URLs and checks whether they exist.
func VerifyStoreMedia(ctx context.Context, service *storefront.Service, opts VerifyStoreMediaOptions, appIDs ...uint32) ([]VerifyResult, error) {
	items, err := FetchStoreMediaURLs(ctx, service, StoreMediaOptions{
		CountryCode: opts.CountryCode,
		Language:    opts.Language,
		Kinds:       opts.Kinds,
	}, appIDs...)
	if err != nil {
		return nil, err
	}

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

// DownloadStoreMedia fetches Storefront media URLs and downloads them.
//
// DASH/HLS movie entries download the playlist/manifest URL itself. The helper
// does not expand playlists into video segments.
func DownloadStoreMedia(ctx context.Context, service *storefront.Service, opts DownloadStoreMediaOptions, appIDs ...uint32) ([]DownloadResult, error) {
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

	items, err := FetchStoreMediaURLs(ctx, service, StoreMediaOptions{
		CountryCode: opts.CountryCode,
		Language:    opts.Language,
		Kinds:       opts.Kinds,
	}, appIDs...)
	if err != nil {
		return nil, err
	}

	requests := make([]downloadRequest, 0, len(items))
	usedPaths := make(map[string]int)
	for _, item := range items {
		name := storeMediaFilename(item, opts.FilenameStyle)
		path := uniqueDownloadPath(appDownloadPath(opts.Dir, mode, item.AppID, name), usedPaths)
		requests = append(requests, downloadRequest{
			item: item,
			url:  item.URL,
			path: path,
		})
	}
	return downloadRequests(ctx, opts.HTTPClient, effectiveOverwrite(opts.Overwrite, opts.SkipExisting), opts.Concurrency, requests)
}

func storeMediaItems(appID uint32, data storefront.AppDetailsData, kinds []Kind) []URLItem {
	out := make([]URLItem, 0)
	for _, kind := range kinds {
		switch kind {
		case KindStoreBackground:
			out = appendStoreMediaURL(out, URLItem{AppID: appID, Kind: kind, URL: data.Background})
		case KindStoreBackgroundRaw:
			out = appendStoreMediaURL(out, URLItem{AppID: appID, Kind: kind, URL: data.BackgroundRaw})
		case KindScreenshotThumbnail:
			for _, screenshot := range data.Screenshots {
				out = appendStoreMediaURL(out, URLItem{
					AppID: appID,
					Kind:  kind,
					ID:    screenshot.ID,
					URL:   screenshot.PathThumbnail,
				})
			}
		case KindScreenshotFull:
			for _, screenshot := range data.Screenshots {
				out = appendStoreMediaURL(out, URLItem{
					AppID: appID,
					Kind:  kind,
					ID:    screenshot.ID,
					URL:   screenshot.PathFull,
				})
			}
		case KindMovieThumbnail:
			for _, movie := range data.Movies {
				out = appendStoreMediaURL(out, movieURLItem(appID, kind, movie, movie.Thumbnail))
			}
		case KindMovieWebM480:
			for _, movie := range data.Movies {
				out = appendStoreMediaURL(out, movieURLItem(appID, kind, movie, movie.WebM.P480))
			}
		case KindMovieWebMMax:
			for _, movie := range data.Movies {
				out = appendStoreMediaURL(out, movieURLItem(appID, kind, movie, movie.WebM.Max))
			}
		case KindMovieMP4480:
			for _, movie := range data.Movies {
				out = appendStoreMediaURL(out, movieURLItem(appID, kind, movie, movie.MP4.P480))
			}
		case KindMovieMP4Max:
			for _, movie := range data.Movies {
				out = appendStoreMediaURL(out, movieURLItem(appID, kind, movie, movie.MP4.Max))
			}
		case KindMovieDASHAV1:
			for _, movie := range data.Movies {
				out = appendStoreMediaURL(out, movieURLItem(appID, kind, movie, movie.DASHAV1))
			}
		case KindMovieDASHH264:
			for _, movie := range data.Movies {
				out = appendStoreMediaURL(out, movieURLItem(appID, kind, movie, movie.DASHH264))
			}
		case KindMovieHLSH264:
			for _, movie := range data.Movies {
				out = appendStoreMediaURL(out, movieURLItem(appID, kind, movie, movie.HLSH264))
			}
		}
	}
	return out
}

func movieURLItem(appID uint32, kind Kind, movie storefront.StoreMovie, rawURL string) URLItem {
	return URLItem{
		AppID: appID,
		Kind:  kind,
		ID:    movie.ID,
		Name:  movie.Name,
		URL:   rawURL,
	}
}

func appendStoreMediaURL(items []URLItem, item URLItem) []URLItem {
	if item.URL == "" {
		return items
	}
	return append(items, item)
}

func storeMediaFilename(item URLItem, style FilenameStyle) string {
	if style == "" {
		style = FilenameOriginal
	}
	original, err := httpasset.Filename(item.URL)
	if err != nil {
		original = storeMediaKindFilename(item)
	}
	switch style {
	case FilenameKind:
		return storeMediaKindFilename(item)
	case FilenameAppKind:
		return strconv.FormatUint(uint64(item.AppID), 10) + "_" + storeMediaKindFilename(item)
	default:
		return original
	}
}

func storeMediaKindFilename(item URLItem) string {
	original, err := httpasset.Filename(item.URL)
	ext := ""
	if err == nil {
		ext = filepath.Ext(original)
	}
	if ext == "" {
		switch item.Kind {
		case KindMovieDASHAV1, KindMovieDASHH264:
			ext = ".mpd"
		case KindMovieHLSH264:
			ext = ".m3u8"
		default:
			ext = ".jpg"
		}
	}
	if item.ID != 0 {
		return string(item.Kind) + "_" + strconv.Itoa(item.ID) + ext
	}
	return string(item.Kind) + ext
}

func uniqueDownloadPath(path string, used map[string]int) string {
	count := used[path]
	used[path] = count + 1
	if count == 0 {
		return path
	}
	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]
	return base + "_" + strconv.Itoa(count+1) + ext
}
