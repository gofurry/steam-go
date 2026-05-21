package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/addons/assets"
)

func main() {
	var (
		appIDsRaw          = flag.String("app-ids", "550,107100", "comma-separated Steam app ids")
		countryCode        = flag.String("country-code", "", "optional Storefront country code for Store media")
		language           = flag.String("language", "schinese", "localized header language")
		kind               = flag.String("kind", string(assets.KindLibraryHero), "asset kind for URLs and downloads, or all")
		proxyRaw           = flag.String("proxy", "", "optional HTTP/HTTPS proxy, for example http://127.0.0.1:7897")
		timeout            = flag.Duration("timeout", 30*time.Second, "HTTP timeout for verification/download and Store media requests")
		clientIconHash     = flag.String("client-icon-hash", "", "optional hash for ClientIconURLs demo")
		verifyURLsRaw      = flag.String("verify-urls", "", "optional comma-separated URLs to verify")
		verifyApps         = flag.Bool("verify-apps", false, "verify constructed app asset URLs")
		fetchStoreMedia    = flag.Bool("store-media", false, "fetch Storefront screenshot/movie/background URLs")
		verifyStoreMedia   = flag.Bool("verify-store-media", false, "fetch and verify Storefront screenshot/movie/background URLs")
		readURLsRaw        = flag.String("read-urls", "", "optional comma-separated URLs to read into memory")
		readApps           = flag.Bool("read-apps", false, "read constructed app asset URLs into memory")
		readStoreMedia     = flag.Bool("read-store-media", false, "fetch and read Storefront screenshot/movie/background URLs into memory")
		readMaxBytes       = flag.Int64("read-max-bytes", 32<<20, "maximum bytes to read per resource")
		downloadURLsRaw    = flag.String("download-urls", "", "optional comma-separated URLs to download into -download-dir")
		downloadApps       = flag.Bool("download-apps", false, "download constructed app assets into -download-dir")
		downloadStoreMedia = flag.Bool("download-store-media", false, "fetch and download Storefront screenshot/movie/background URLs")
		downloadDir        = flag.String("download-dir", "", "optional directory for downloads")
		downloadMode       = flag.String("download-mode", string(assets.StoreFlat), "app download mode: flat or by_app_id")
		filenameStyle      = flag.String("filename-style", string(assets.FilenameOriginal), "app filename style: original, kind, or app_kind")
		overwrite          = flag.String("overwrite", string(assets.OverwriteAlways), "overwrite mode: always, never, or if_zero")
		skipExisting       = flag.Bool("skip-existing", false, "skip existing downloaded files")
		concurrency        = flag.Int("concurrency", 1, "number of concurrent downloads")
		manifestPath       = flag.String("manifest", "", "optional path to write a JSON manifest")
	)
	flag.Parse()

	appIDs, err := parseAppIDs(*appIDsRaw)
	if err != nil {
		log.Fatal(err)
	}
	selector, httpClient, err := proxyHTTPClient(*proxyRaw, *timeout)
	if err != nil {
		log.Fatal(err)
	}

	out := exampleOutput{
		Input: exampleInput{
			AppIDs:             appIDs,
			CountryCode:        *countryCode,
			Language:           *language,
			Kind:               assets.Kind(*kind),
			Proxy:              *proxyRaw,
			Timeout:            timeout.String(),
			ClientIconHash:     *clientIconHash,
			VerifyURLs:         splitCSV(*verifyURLsRaw),
			VerifyApps:         *verifyApps,
			StoreMedia:         *fetchStoreMedia,
			VerifyStoreMedia:   *verifyStoreMedia,
			ReadURLs:           splitCSV(*readURLsRaw),
			ReadApps:           *readApps,
			ReadStoreMedia:     *readStoreMedia,
			ReadMaxBytes:       *readMaxBytes,
			DownloadURLs:       splitCSV(*downloadURLsRaw),
			DownloadApps:       *downloadApps,
			DownloadStoreMedia: *downloadStoreMedia,
			DownloadDir:        *downloadDir,
			DownloadMode:       assets.StoreMode(*downloadMode),
			FilenameStyle:      assets.FilenameStyle(*filenameStyle),
			Overwrite:          assets.OverwriteMode(*overwrite),
			SkipExisting:       *skipExisting,
			Concurrency:        *concurrency,
			ManifestPath:       *manifestPath,
		},

		SingleResourceHelpers: singleResourceOutput{
			HeaderURLs:          assets.HeaderURLs(appIDs...),
			HeaderLocalizedURLs: assets.HeaderLocalizedURLs(*language, appIDs...),
			CapsuleSmallURLs:    assets.CapsuleSmallURLs(appIDs...),
			CapsuleMainURLs:     assets.CapsuleMainURLs(appIDs...),
			LibraryCapsuleURLs:  assets.LibraryCapsuleURLs(appIDs...),
			LibraryCapsule2xURLs: assets.LibraryCapsule2xURLs(
				appIDs...,
			),
			LibraryHeroURLs:   assets.LibraryHeroURLs(appIDs...),
			LibraryLogoURLs:   assets.LibraryLogoURLs(appIDs...),
			LibraryLogo2xURLs: assets.LibraryLogo2xURLs(appIDs...),
		},

		ResourceList: resourceList(assets.Kind(*kind), *language, appIDs),
		URLsByKind:   urlsByKind(assets.Kind(*kind), appIDs),
		AllByAppID: assets.AllWithLanguage(
			*language,
			appIDs...,
		),
	}

	if *clientIconHash != "" && len(appIDs) > 0 {
		out.ClientIconURLs = assets.ClientIconURLs(assets.HashRef{
			AppID: appIDs[0],
			Hash:  *clientIconHash,
		})
	}

	verifyURLs := splitCSV(*verifyURLsRaw)
	if len(verifyURLs) > 0 {
		out.VerifyResults, err = assets.VerifyURLsWithClient(context.Background(), httpClient, verifyURLs...)
		if err != nil {
			log.Fatal(err)
		}
	}
	if *verifyApps {
		out.VerifyAppResults, err = assets.VerifyAppAssets(context.Background(), assets.VerifyAppOptions{
			Kinds:      downloadKinds(*kind),
			Language:   *language,
			HTTPClient: httpClient,
		}, appIDs...)
		if err != nil {
			log.Fatal(err)
		}
	}
	readURLs := splitCSV(*readURLsRaw)
	if len(readURLs) > 0 {
		var readErr error
		readResults, readErr := assets.ReadURLsWithOptions(context.Background(), assets.ReadOptions{
			HTTPClient:  httpClient,
			MaxBytes:    *readMaxBytes,
			Concurrency: *concurrency,
		}, readURLs...)
		out.ReadURLResults = summarizeReadResults(readResults)
		out.ReadURLError = errorString(readErr)
	}
	if *readApps {
		var readErr error
		readResults, readErr := assets.ReadAppAssets(context.Background(), assets.ReadAppOptions{
			Kinds:       downloadKinds(*kind),
			Language:    *language,
			HTTPClient:  httpClient,
			MaxBytes:    *readMaxBytes,
			Concurrency: *concurrency,
		}, appIDs...)
		out.ReadAppResults = summarizeReadResults(readResults)
		out.ReadAppError = errorString(readErr)
	}
	if *fetchStoreMedia || *verifyStoreMedia || *readStoreMedia || *downloadStoreMedia {
		sdkOpts := []steam.Option{steam.WithTimeout(*timeout)}
		if selector != nil {
			sdkOpts = append(sdkOpts, steam.WithProxySelector(selector))
		}
		sdk, err := steam.NewClient(sdkOpts...)
		if err != nil {
			log.Fatal(err)
		}
		if *fetchStoreMedia {
			out.StoreMediaURLs, err = assets.FetchStoreMediaURLs(context.Background(), sdk.Web.Storefront, assets.StoreMediaOptions{
				CountryCode: *countryCode,
				Language:    *language,
				Kinds:       storeMediaKinds(*kind),
			}, appIDs...)
			if err != nil {
				log.Fatal(err)
			}
		}
		if *verifyStoreMedia {
			out.VerifyStoreMediaResults, err = assets.VerifyStoreMedia(context.Background(), sdk.Web.Storefront, assets.VerifyStoreMediaOptions{
				CountryCode: *countryCode,
				Language:    *language,
				Kinds:       storeMediaKinds(*kind),
				HTTPClient:  httpClient,
			}, appIDs...)
			if err != nil {
				log.Fatal(err)
			}
		}
		if *readStoreMedia {
			var readErr error
			readResults, readErr := assets.ReadStoreMedia(context.Background(), sdk.Web.Storefront, assets.ReadStoreMediaOptions{
				CountryCode: *countryCode,
				Language:    *language,
				Kinds:       storeMediaKinds(*kind),
				HTTPClient:  httpClient,
				MaxBytes:    *readMaxBytes,
				Concurrency: *concurrency,
			}, appIDs...)
			out.ReadStoreMediaResults = summarizeReadResults(readResults)
			out.ReadStoreMediaError = errorString(readErr)
		}
		if *downloadStoreMedia {
			if *downloadDir == "" {
				log.Fatal("-download-dir is required with -download-store-media")
			}
			var downloadErr error
			out.DownloadStoreMediaResults, downloadErr = assets.DownloadStoreMedia(context.Background(), sdk.Web.Storefront, assets.DownloadStoreMediaOptions{
				Dir:           *downloadDir,
				CountryCode:   *countryCode,
				Language:      *language,
				Kinds:         storeMediaKinds(*kind),
				Mode:          assets.StoreMode(*downloadMode),
				HTTPClient:    httpClient,
				Overwrite:     assets.OverwriteMode(*overwrite),
				SkipExisting:  *skipExisting,
				FilenameStyle: assets.FilenameStyle(*filenameStyle),
				Concurrency:   *concurrency,
			}, appIDs...)
			out.DownloadStoreMediaError = errorString(downloadErr)
		}
	}

	downloadURLs := splitCSV(*downloadURLsRaw)
	if len(downloadURLs) > 0 {
		if *downloadDir == "" {
			log.Fatal("-download-dir is required with -download-urls")
		}
		var downloadErr error
		out.DownloadURLResults, downloadErr = assets.DownloadURLsWithOptions(context.Background(), assets.DownloadOptions{
			Dir:          *downloadDir,
			HTTPClient:   httpClient,
			Overwrite:    assets.OverwriteMode(*overwrite),
			SkipExisting: *skipExisting,
			Concurrency:  *concurrency,
		}, downloadURLs...)
		out.DownloadURLError = errorString(downloadErr)
	}

	if *downloadApps {
		if *downloadDir == "" {
			log.Fatal("-download-dir is required with -download-apps")
		}
		var downloadErr error
		out.DownloadAppResults, downloadErr = assets.DownloadAppAssets(context.Background(), assets.DownloadAppOptions{
			Dir:           *downloadDir,
			Kinds:         downloadKinds(*kind),
			Language:      *language,
			Mode:          assets.StoreMode(*downloadMode),
			HTTPClient:    httpClient,
			Overwrite:     assets.OverwriteMode(*overwrite),
			SkipExisting:  *skipExisting,
			FilenameStyle: assets.FilenameStyle(*filenameStyle),
			Concurrency:   *concurrency,
		}, appIDs...)
		out.DownloadAppError = errorString(downloadErr)
	}

	if *manifestPath != "" {
		manifest := assets.NewURLManifest(out.ResourceList)
		if len(out.StoreMediaURLs) > 0 {
			manifest = assets.NewURLManifest(append(out.ResourceList, out.StoreMediaURLs...))
		}
		if len(out.DownloadAppResults) > 0 || len(out.DownloadURLResults) > 0 || len(out.DownloadStoreMediaResults) > 0 {
			downloads := append(out.DownloadURLResults, out.DownloadAppResults...)
			downloads = append(downloads, out.DownloadStoreMediaResults...)
			manifest = assets.NewDownloadManifest(downloads)
		}
		if err := assets.WriteManifestJSON(*manifestPath, manifest); err != nil {
			log.Fatal(err)
		}
		out.ManifestPath = *manifestPath
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(out); err != nil {
		log.Fatal(err)
	}
}

type exampleInput struct {
	AppIDs             []uint32             `json:"app_ids"`
	CountryCode        string               `json:"country_code,omitempty"`
	Language           string               `json:"language"`
	Kind               assets.Kind          `json:"kind"`
	Proxy              string               `json:"proxy,omitempty"`
	Timeout            string               `json:"timeout,omitempty"`
	ClientIconHash     string               `json:"client_icon_hash,omitempty"`
	VerifyURLs         []string             `json:"verify_urls,omitempty"`
	VerifyApps         bool                 `json:"verify_apps,omitempty"`
	StoreMedia         bool                 `json:"store_media,omitempty"`
	VerifyStoreMedia   bool                 `json:"verify_store_media,omitempty"`
	ReadURLs           []string             `json:"read_urls,omitempty"`
	ReadApps           bool                 `json:"read_apps,omitempty"`
	ReadStoreMedia     bool                 `json:"read_store_media,omitempty"`
	ReadMaxBytes       int64                `json:"read_max_bytes,omitempty"`
	DownloadURLs       []string             `json:"download_urls,omitempty"`
	DownloadApps       bool                 `json:"download_apps,omitempty"`
	DownloadStoreMedia bool                 `json:"download_store_media,omitempty"`
	DownloadDir        string               `json:"download_dir,omitempty"`
	DownloadMode       assets.StoreMode     `json:"download_mode,omitempty"`
	FilenameStyle      assets.FilenameStyle `json:"filename_style,omitempty"`
	Overwrite          assets.OverwriteMode `json:"overwrite,omitempty"`
	SkipExisting       bool                 `json:"skip_existing,omitempty"`
	Concurrency        int                  `json:"concurrency,omitempty"`
	ManifestPath       string               `json:"manifest_path,omitempty"`
}

type singleResourceOutput struct {
	HeaderURLs           []string `json:"header_urls"`
	HeaderLocalizedURLs  []string `json:"header_localized_urls"`
	CapsuleSmallURLs     []string `json:"capsule_small_urls"`
	CapsuleMainURLs      []string `json:"capsule_main_urls"`
	LibraryCapsuleURLs   []string `json:"library_capsule_urls"`
	LibraryCapsule2xURLs []string `json:"library_capsule_2x_urls"`
	LibraryHeroURLs      []string `json:"library_hero_urls"`
	LibraryLogoURLs      []string `json:"library_logo_urls"`
	LibraryLogo2xURLs    []string `json:"library_logo_2x_urls"`
}

type exampleOutput struct {
	Input                     exampleInput            `json:"input"`
	SingleResourceHelpers     singleResourceOutput    `json:"single_resource_helpers"`
	ResourceList              []assets.URLItem        `json:"resource_list"`
	URLsByKind                []string                `json:"urls_by_kind"`
	AllByAppID                []assets.AppAssets      `json:"all_by_app_id"`
	ClientIconURLs            []string                `json:"client_icon_urls,omitempty"`
	VerifyResults             []assets.VerifyResult   `json:"verify_results,omitempty"`
	VerifyAppResults          []assets.VerifyResult   `json:"verify_app_results,omitempty"`
	StoreMediaURLs            []assets.URLItem        `json:"store_media_urls,omitempty"`
	VerifyStoreMediaResults   []assets.VerifyResult   `json:"verify_store_media_results,omitempty"`
	ReadURLResults            []readResultSummary     `json:"read_url_results,omitempty"`
	ReadURLError              string                  `json:"read_url_error,omitempty"`
	ReadAppResults            []readResultSummary     `json:"read_app_results,omitempty"`
	ReadAppError              string                  `json:"read_app_error,omitempty"`
	ReadStoreMediaResults     []readResultSummary     `json:"read_store_media_results,omitempty"`
	ReadStoreMediaError       string                  `json:"read_store_media_error,omitempty"`
	DownloadURLResults        []assets.DownloadResult `json:"download_url_results,omitempty"`
	DownloadURLError          string                  `json:"download_url_error,omitempty"`
	DownloadAppResults        []assets.DownloadResult `json:"download_app_results,omitempty"`
	DownloadAppError          string                  `json:"download_app_error,omitempty"`
	DownloadStoreMediaResults []assets.DownloadResult `json:"download_store_media_results,omitempty"`
	DownloadStoreMediaError   string                  `json:"download_store_media_error,omitempty"`
	ManifestPath              string                  `json:"manifest_path,omitempty"`
}

type readResultSummary struct {
	AppID             uint32      `json:"app_id,omitempty"`
	Kind              assets.Kind `json:"kind,omitempty"`
	ID                int         `json:"id,omitempty"`
	Name              string      `json:"name,omitempty"`
	URL               string      `json:"url"`
	StatusCode        int         `json:"status_code,omitempty"`
	ContentType       string      `json:"content_type,omitempty"`
	ContentLength     int64       `json:"content_length,omitempty"`
	BytesRead         int64       `json:"bytes_read,omitempty"`
	DataPreviewBase64 string      `json:"data_preview_base64,omitempty"`
	Error             string      `json:"error,omitempty"`
}

func parseAppIDs(raw string) ([]uint32, error) {
	parts := strings.Split(raw, ",")
	out := make([]uint32, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		value, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, err
		}
		out = append(out, uint32(value))
	}
	return out, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func downloadKinds(raw string) []assets.Kind {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "all" {
		return nil
	}
	return []assets.Kind{assets.Kind(raw)}
}

func storeMediaKinds(raw string) []assets.Kind {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "all" {
		return nil
	}
	return []assets.Kind{assets.Kind(raw)}
}

func proxyHTTPClient(proxyRaw string, timeout time.Duration) (steam.ProxySelector, *http.Client, error) {
	selector, err := steam.NewStaticProxySelector(proxyRaw)
	if err != nil {
		return nil, nil, err
	}
	client, err := steam.NewHTTPClientWithProxySelector(selector, timeout)
	if err != nil {
		return nil, nil, err
	}
	return selector, client, nil
}

func summarizeReadResults(results []assets.ReadResult) []readResultSummary {
	out := make([]readResultSummary, 0, len(results))
	for _, result := range results {
		preview := result.Data
		if len(preview) > 64 {
			preview = preview[:64]
		}
		out = append(out, readResultSummary{
			AppID:             result.AppID,
			Kind:              result.Kind,
			ID:                result.ID,
			Name:              result.Name,
			URL:               result.URL,
			StatusCode:        result.StatusCode,
			ContentType:       result.ContentType,
			ContentLength:     result.ContentLength,
			BytesRead:         result.BytesRead,
			DataPreviewBase64: base64.StdEncoding.EncodeToString(preview),
			Error:             result.Error,
		})
	}
	return out
}

func resourceList(kind assets.Kind, language string, appIDs []uint32) []assets.URLItem {
	if kind == "all" {
		return assets.ListWithLanguage(language, appIDs...)
	}
	return assets.ListKindsWithLanguage(language, []assets.Kind{kind}, appIDs...)
}

func urlsByKind(kind assets.Kind, appIDs []uint32) []string {
	if kind == "all" {
		return nil
	}
	return assets.URLs(kind, appIDs...)
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
