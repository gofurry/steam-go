package assets

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofurry/steam-go/addons/assets/internal/httpasset"
	"github.com/gofurry/steam-go/api/storebrowseservice"
)

const DefaultStoreItemAssetBaseURL = "https://shared.steamstatic.com/store_item_assets/"

var storeItemAssetPathRE = regexp.MustCompile(`/steam/apps/\d+/([0-9a-fA-F]{40})/([^/?#]+)`)

// FetchStoreItemAssetURLs requests official StoreBrowse item asset metadata and
// returns resolved asset URLs. Results follow appID order, then requested kind
// order. Missing asset fields are skipped.
//
// If opts.Kinds is empty, StoreItemAssetKinds is used.
func FetchStoreItemAssetURLs(ctx context.Context, service *storebrowseservice.Service, opts StoreItemAssetOptions, appIDs ...uint32) ([]URLItem, error) {
	if service == nil {
		return nil, fmt.Errorf("store browse service must not be nil")
	}
	ids := make([]storebrowseservice.StoreItemID, 0, len(appIDs))
	for _, appID := range appIDs {
		if appID == 0 {
			return nil, fmt.Errorf("app id must be greater than zero")
		}
		ids = append(ids, storebrowseservice.StoreItemID{AppID: appID})
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("app id must be greater than zero")
	}

	req := storebrowseservice.GetItemsRequest{
		IDs: ids,
		DataRequest: &storebrowseservice.StoreBrowseDataRequest{
			IncludeAssets: true,
		},
	}
	if strings.TrimSpace(opts.CountryCode) != "" || strings.TrimSpace(opts.Language) != "" {
		req.Context = &storebrowseservice.StoreBrowseContext{
			CountryCode: strings.TrimSpace(opts.CountryCode),
			Language:    strings.TrimSpace(opts.Language),
		}
	}

	resp, err := service.GetItems(ctx, req)
	if err != nil {
		return nil, err
	}

	itemsByAppID := make(map[uint32]storebrowseservice.StoreItem, len(resp.Response.StoreItems))
	for _, item := range resp.Response.StoreItems {
		appID := storeItemAppID(item)
		if appID == 0 {
			continue
		}
		itemsByAppID[appID] = item
	}

	kinds := opts.Kinds
	if len(kinds) == 0 {
		kinds = StoreItemAssetKinds()
	}
	baseURL := opts.BaseURL
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultStoreItemAssetBaseURL
	}

	out := make([]URLItem, 0)
	for _, appID := range appIDs {
		item, ok := itemsByAppID[appID]
		if !ok {
			continue
		}
		out = append(out, storeItemAssetItems(appID, item, kinds, baseURL, opts.StripQuery)...)
	}
	return out, nil
}

// VerifyStoreItemAssets fetches official Store item asset URLs and checks
// whether each URL resolves to a 2xx HTTP response.
func VerifyStoreItemAssets(ctx context.Context, service *storebrowseservice.Service, opts VerifyStoreItemAssetOptions, appIDs ...uint32) ([]VerifyResult, error) {
	items, err := FetchStoreItemAssetURLs(ctx, service, StoreItemAssetOptions{
		CountryCode: opts.CountryCode,
		Language:    opts.Language,
		Kinds:       opts.Kinds,
		BaseURL:     opts.BaseURL,
		StripQuery:  opts.StripQuery,
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

// ReadStoreItemAssets fetches official Store item asset URLs and reads them
// into memory.
func ReadStoreItemAssets(ctx context.Context, service *storebrowseservice.Service, opts ReadStoreItemAssetOptions, appIDs ...uint32) ([]ReadResult, error) {
	items, err := FetchStoreItemAssetURLs(ctx, service, StoreItemAssetOptions{
		CountryCode: opts.CountryCode,
		Language:    opts.Language,
		Kinds:       opts.Kinds,
		BaseURL:     opts.BaseURL,
		StripQuery:  opts.StripQuery,
	}, appIDs...)
	if err != nil {
		return nil, err
	}
	return readURLItems(ctx, opts.HTTPClient, opts.MaxBytes, opts.Concurrency, items)
}

// DownloadStoreItemAssets fetches official Store item asset URLs and downloads
// them into opts.Dir.
func DownloadStoreItemAssets(ctx context.Context, service *storebrowseservice.Service, opts DownloadStoreItemAssetOptions, appIDs ...uint32) ([]DownloadResult, error) {
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

	items, err := FetchStoreItemAssetURLs(ctx, service, StoreItemAssetOptions{
		CountryCode: opts.CountryCode,
		Language:    opts.Language,
		Kinds:       opts.Kinds,
		BaseURL:     opts.BaseURL,
		StripQuery:  opts.StripQuery,
	}, appIDs...)
	if err != nil {
		return nil, err
	}

	requests := make([]downloadRequest, 0, len(items))
	usedPaths := make(map[string]int)
	for _, item := range items {
		name := storeItemAssetFilename(item, opts.FilenameStyle)
		path := uniqueDownloadPath(appDownloadPath(opts.Dir, mode, item.AppID, name), usedPaths)
		requests = append(requests, downloadRequest{
			item: item,
			url:  item.URL,
			path: path,
		})
	}
	return downloadRequests(ctx, opts.HTTPClient, effectiveOverwrite(opts.Overwrite, opts.SkipExisting), opts.Concurrency, requests)
}

// ResolveStoreItemAssetURL resolves one StoreBrowse asset filename with the
// supplied asset_url_format and base URL.
func ResolveStoreItemAssetURL(baseURL, format, filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return ""
	}
	if isAbsoluteHTTPURL(filename) {
		return filename
	}

	format = strings.TrimSpace(format)
	if format == "" {
		return joinStoreItemAssetURL(baseURL, filename)
	}

	cleanFilename := strings.TrimLeft(filename, "/")
	var path string
	switch {
	case strings.Contains(format, "${FILENAME}"):
		path = strings.ReplaceAll(format, "${FILENAME}", cleanFilename)
	case strings.Contains(format, "%s"):
		path = fmt.Sprintf(format, cleanFilename)
	default:
		path = strings.TrimRight(format, "/") + "/" + cleanFilename
	}
	if isAbsoluteHTTPURL(path) {
		return path
	}
	return joinStoreItemAssetURL(baseURL, path)
}

// ParseStoreItemAssetURL extracts the digest and filename from one hashed Store
// item asset URL.
func ParseStoreItemAssetURL(rawURL string) (digest, filename string) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", ""
	}
	matches := storeItemAssetPathRE.FindStringSubmatch(parsed.Path)
	if len(matches) != 3 {
		return "", ""
	}
	return matches[1], matches[2]
}

func storeItemAssetItems(appID uint32, item storebrowseservice.StoreItem, kinds []Kind, baseURL string, stripQuery bool) []URLItem {
	format := item.Assets["asset_url_format"]
	out := make([]URLItem, 0, len(kinds))
	for _, kind := range kinds {
		if kind == KindCommunityIconJPG {
			out = appendCommunityIconItem(out, appID, item)
			continue
		}

		filename := firstStoreItemAssetValue(item.Assets, storeItemAssetKeys(kind))
		if strings.TrimSpace(filename) == "" {
			continue
		}
		rawURL := ResolveStoreItemAssetURL(baseURL, format, filename)
		if rawURL == "" {
			continue
		}
		if stripQuery {
			rawURL = stripRawQuery(rawURL)
		}
		digest, parsedFilename := ParseStoreItemAssetURL(rawURL)
		if parsedFilename == "" {
			parsedFilename = filepath.Base(strings.TrimSpace(filename))
		}
		out = append(out, URLItem{
			AppID:    appID,
			Kind:     kind,
			Name:     item.Name,
			URL:      rawURL,
			Digest:   digest,
			Filename: parsedFilename,
			Source:   SourceStoreBrowse,
		})
	}
	return out
}

func appendCommunityIconItem(items []URLItem, appID uint32, item storebrowseservice.StoreItem) []URLItem {
	hash := strings.TrimSpace(item.Assets["community_icon"])
	if hash == "" {
		return items
	}
	rawURL := CommunityIconURL(appID, hash)
	if rawURL == "" {
		return items
	}
	return append(items, URLItem{
		AppID:    appID,
		Kind:     KindCommunityIconJPG,
		Name:     item.Name,
		URL:      rawURL,
		Digest:   hash,
		Filename: "community_icon.jpg",
		Source:   SourceStoreBrowse,
	})
}

func storeItemAssetKeys(kind Kind) []string {
	switch kind {
	case KindCapsuleMain:
		return []string{"main_capsule"}
	case KindCapsuleMain2x:
		return []string{"main_capsule_2x"}
	case KindCapsuleSmall:
		return []string{"small_capsule"}
	case KindCapsuleSmall2x:
		return []string{"small_capsule_2x"}
	case KindHeader:
		return []string{"header"}
	case KindHeader2x:
		return []string{"header_2x"}
	case KindHeroCapsule:
		return []string{"hero_capsule"}
	case KindHeroCapsule2x:
		return []string{"hero_capsule_2x"}
	case KindLibraryCapsule:
		return []string{"library_capsule"}
	case KindLibraryCapsule2x:
		return []string{"library_capsule_2x"}
	case KindLibraryHero:
		return []string{"library_hero"}
	case KindLibraryHero2x:
		return []string{"library_hero_2x"}
	case KindLibraryLogo:
		return []string{"library_logo", "logo"}
	case KindLibraryLogo2x:
		return []string{"library_logo_2x", "logo_2x"}
	case KindPageBackground:
		return []string{"page_background"}
	case KindPageBackgroundRaw:
		return []string{"raw_page_background"}
	default:
		return nil
	}
}

func firstStoreItemAssetValue(assets storebrowseservice.StoreItemAssets, keys []string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(assets[key]); value != "" {
			return value
		}
	}
	return ""
}

func storeItemAppID(item storebrowseservice.StoreItem) uint32 {
	if item.AppID != 0 {
		return item.AppID
	}
	return item.ID
}

func joinStoreItemAssetURL(baseURL, path string) string {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultStoreItemAssetBaseURL
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
}

func isAbsoluteHTTPURL(rawURL string) bool {
	return strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://")
}

func stripRawQuery(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	return parsed.String()
}

func storeItemAssetFilename(item URLItem, style FilenameStyle) string {
	if style == "" {
		style = FilenameOriginal
	}
	original := strings.TrimSpace(item.Filename)
	if original == "" {
		if name, err := httpasset.Filename(item.URL); err == nil {
			original = name
		}
	}
	if original == "" {
		original = storeItemAssetKindFilename(item)
	}

	switch style {
	case FilenameKind:
		return storeItemAssetKindFilename(item)
	case FilenameAppKind:
		return strconv.FormatUint(uint64(item.AppID), 10) + "_" + storeItemAssetKindFilename(item)
	default:
		return original
	}
}

func storeItemAssetKindFilename(item URLItem) string {
	ext := filepath.Ext(strings.TrimSpace(item.Filename))
	if ext == "" {
		if original, err := httpasset.Filename(item.URL); err == nil {
			ext = filepath.Ext(original)
		}
	}
	if ext == "" {
		ext = ".jpg"
	}
	return string(item.Kind) + ext
}
