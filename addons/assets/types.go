package assets

import "net/http"

// Kind identifies one Steam asset family.
type Kind string

const (
	KindHeader           Kind = "header"
	KindHeaderLocalized  Kind = "header_localized"
	KindHeader2x         Kind = "header_2x"
	KindCapsuleSmall     Kind = "capsule_small"
	KindCapsuleSmall2x   Kind = "capsule_small_2x"
	KindCapsuleMain      Kind = "capsule_main"
	KindCapsuleMain2x    Kind = "capsule_main_2x"
	KindHeroCapsule      Kind = "hero_capsule"
	KindHeroCapsule2x    Kind = "hero_capsule_2x"
	KindLibraryCapsule   Kind = "library_capsule"
	KindLibraryCapsule2x Kind = "library_capsule_2x"
	KindLibraryHero      Kind = "library_hero"
	KindLibraryHero2x    Kind = "library_hero_2x"
	KindLibraryLogo      Kind = "library_logo"
	KindLibraryLogo2x    Kind = "library_logo_2x"
	KindCommunityIconJPG Kind = "community_icon_jpg"
	KindCommunityLogoJPG Kind = "community_logo_jpg"
	KindClientIconICO    Kind = "client_icon_ico"

	KindStoreBackground     Kind = "store_background"
	KindStoreBackgroundRaw  Kind = "store_background_raw"
	KindPageBackground      Kind = "page_background"
	KindPageBackgroundRaw   Kind = "page_background_raw"
	KindScreenshotThumbnail Kind = "screenshot_thumbnail"
	KindScreenshotFull      Kind = "screenshot_full"
	KindMovieThumbnail      Kind = "movie_thumbnail"
	KindMovieWebM480        Kind = "movie_webm_480"
	KindMovieWebMMax        Kind = "movie_webm_max"
	KindMovieMP4480         Kind = "movie_mp4_480"
	KindMovieMP4Max         Kind = "movie_mp4_max"
	KindMovieDASHAV1        Kind = "movie_dash_av1"
	KindMovieDASHH264       Kind = "movie_dash_h264"
	KindMovieHLSH264        Kind = "movie_hls_h264"
)

const (
	// SourceLegacyStatic identifies locally constructed AppID-only asset URLs.
	SourceLegacyStatic = "legacy_static"

	// SourceStoreBrowse identifies URLs discovered from IStoreBrowseService.
	SourceStoreBrowse = "store_browse"

	// SourceStorefrontAppDetails identifies URLs discovered from Storefront appdetails.
	SourceStorefrontAppDetails = "storefront_appdetails"
)

// URLItem describes one constructed asset URL.
type URLItem struct {
	AppID    uint32 `json:"app_id,omitempty"`
	Kind     Kind   `json:"kind,omitempty"`
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	URL      string `json:"url"`
	Digest   string `json:"digest,omitempty"`
	Filename string `json:"filename,omitempty"`
	Source   string `json:"source,omitempty"`
}

// AppAssets contains the standard Store and Library static asset URLs for one AppID.
//
// HeaderLocalized is filled only by AllWithLanguage.
type AppAssets struct {
	AppID            uint32 `json:"app_id"`
	Header           string `json:"header"`
	HeaderLocalized  string `json:"header_localized,omitempty"`
	CapsuleSmall     string `json:"capsule_small"`
	CapsuleMain      string `json:"capsule_main"`
	LibraryCapsule   string `json:"library_capsule"`
	LibraryCapsule2x string `json:"library_capsule_2x"`
	LibraryHero      string `json:"library_hero"`
	LibraryLogo      string `json:"library_logo"`
	LibraryLogo2x    string `json:"library_logo_2x"`
}

// HashRef pairs an AppID with a Steam image hash for community/client icon URLs.
type HashRef struct {
	AppID uint32 `json:"app_id"`
	Hash  string `json:"hash"`
}

// StoreMode controls where app asset downloads are written.
type StoreMode string

const (
	// StoreFlat writes all app asset files directly under the destination directory.
	//
	// Generated app asset filenames are prefixed with the AppID to avoid collisions,
	// for example "550_header.jpg".
	StoreFlat StoreMode = "flat"

	// StoreByAppID writes each app's assets under a child directory named by AppID.
	StoreByAppID StoreMode = "by_app_id"
)

// FilenameStyle controls generated file names for app asset downloads.
type FilenameStyle string

const (
	// FilenameOriginal keeps Steam's original file name, for example "header.jpg".
	FilenameOriginal FilenameStyle = "original"

	// FilenameKind uses the asset kind as the file name, for example "library_hero.jpg".
	FilenameKind FilenameStyle = "kind"

	// FilenameAppKind prefixes the asset kind with AppID, for example "550_library_hero.jpg".
	FilenameAppKind FilenameStyle = "app_kind"
)

// OverwriteMode controls what happens when the destination file already exists.
type OverwriteMode string

const (
	// OverwriteAlways replaces existing files.
	OverwriteAlways OverwriteMode = "always"

	// OverwriteNever skips existing files.
	OverwriteNever OverwriteMode = "never"

	// OverwriteIfZero replaces only existing empty files.
	OverwriteIfZero OverwriteMode = "if_zero"
)

// DownloadStatus reports the outcome for one download item.
type DownloadStatus string

const (
	DownloadStatusDownloaded DownloadStatus = "downloaded"
	DownloadStatusSkipped    DownloadStatus = "skipped"
	DownloadStatusFailed     DownloadStatus = "failed"
)

// VerifyResult is the result of checking one URL.
type VerifyResult struct {
	AppID         uint32 `json:"app_id,omitempty"`
	Kind          Kind   `json:"kind,omitempty"`
	ID            int    `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	URL           string `json:"url"`
	Digest        string `json:"digest,omitempty"`
	Filename      string `json:"filename,omitempty"`
	Source        string `json:"source,omitempty"`
	Exists        bool   `json:"exists"`
	StatusCode    int    `json:"status_code,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	ContentLength int64  `json:"content_length,omitempty"`
}

// DownloadResult is the result of saving one URL to disk.
type DownloadResult struct {
	AppID         uint32         `json:"app_id,omitempty"`
	Kind          Kind           `json:"kind,omitempty"`
	ID            int            `json:"id,omitempty"`
	Name          string         `json:"name,omitempty"`
	URL           string         `json:"url"`
	Digest        string         `json:"digest,omitempty"`
	Filename      string         `json:"filename,omitempty"`
	Source        string         `json:"source,omitempty"`
	Path          string         `json:"path"`
	Status        DownloadStatus `json:"status"`
	StatusCode    int            `json:"status_code,omitempty"`
	ContentType   string         `json:"content_type,omitempty"`
	ContentLength int64          `json:"content_length,omitempty"`
	BytesWritten  int64          `json:"bytes_written,omitempty"`
	Error         string         `json:"error,omitempty"`
}

// ReadResult is the result of reading one URL into memory.
type ReadResult struct {
	AppID         uint32 `json:"app_id,omitempty"`
	Kind          Kind   `json:"kind,omitempty"`
	ID            int    `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	URL           string `json:"url"`
	Digest        string `json:"digest,omitempty"`
	Filename      string `json:"filename,omitempty"`
	Source        string `json:"source,omitempty"`
	StatusCode    int    `json:"status_code,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	ContentLength int64  `json:"content_length,omitempty"`
	BytesRead     int64  `json:"bytes_read,omitempty"`
	Data          []byte `json:"data,omitempty"`
	Error         string `json:"error,omitempty"`
}

// ReadHandler handles one read result without requiring the whole batch to be
// retained in memory. The Data field is valid for the duration of the call; copy
// it if it must outlive the handler.
type ReadHandler func(ReadResult) error

// ReadOptions controls direct URL reads.
type ReadOptions struct {
	HTTPClient   *http.Client
	MaxBytes     int64
	Concurrency  int
	URLValidator URLValidator
}

// ReadAppOptions controls app asset reads.
type ReadAppOptions struct {
	Kinds       []Kind
	Language    string
	HTTPClient  *http.Client
	MaxBytes    int64
	Concurrency int
}

// DownloadOptions controls direct URL downloads.
type DownloadOptions struct {
	Dir           string
	HTTPClient    *http.Client
	Overwrite     OverwriteMode
	SkipExisting  bool
	FilenameStyle FilenameStyle
	Concurrency   int
	URLValidator  URLValidator
}

// VerifyOptions controls direct URL verification.
type VerifyOptions struct {
	HTTPClient   *http.Client
	URLValidator URLValidator
}

// DownloadAppOptions controls app asset downloads.
type DownloadAppOptions struct {
	Dir           string
	Kinds         []Kind
	Language      string
	Mode          StoreMode
	HTTPClient    *http.Client
	Overwrite     OverwriteMode
	SkipExisting  bool
	FilenameStyle FilenameStyle
	Concurrency   int
}

// VerifyAppOptions controls app asset verification.
type VerifyAppOptions struct {
	Kinds      []Kind
	Language   string
	HTTPClient *http.Client
}

// StoreMediaOptions controls Storefront-backed media URL discovery.
type StoreMediaOptions struct {
	CountryCode string
	Language    string
	Kinds       []Kind
}

// StoreItemAssetOptions controls StoreBrowse-backed Store item asset discovery.
type StoreItemAssetOptions struct {
	CountryCode string
	Language    string
	Kinds       []Kind
	BaseURL     string
	StripQuery  bool
}

// VerifyStoreMediaOptions controls Storefront-backed media URL verification.
type VerifyStoreMediaOptions struct {
	CountryCode string
	Language    string
	Kinds       []Kind
	HTTPClient  *http.Client
}

// VerifyStoreItemAssetOptions controls StoreBrowse-backed asset verification.
type VerifyStoreItemAssetOptions struct {
	CountryCode string
	Language    string
	Kinds       []Kind
	BaseURL     string
	StripQuery  bool
	HTTPClient  *http.Client
}

// DownloadStoreMediaOptions controls Storefront-backed media downloads.
type DownloadStoreMediaOptions struct {
	Dir           string
	CountryCode   string
	Language      string
	Kinds         []Kind
	Mode          StoreMode
	HTTPClient    *http.Client
	Overwrite     OverwriteMode
	SkipExisting  bool
	FilenameStyle FilenameStyle
	Concurrency   int
}

// DownloadStoreItemAssetOptions controls StoreBrowse-backed asset downloads.
type DownloadStoreItemAssetOptions struct {
	Dir           string
	CountryCode   string
	Language      string
	Kinds         []Kind
	BaseURL       string
	StripQuery    bool
	Mode          StoreMode
	HTTPClient    *http.Client
	Overwrite     OverwriteMode
	SkipExisting  bool
	FilenameStyle FilenameStyle
	Concurrency   int
}

// ReadStoreMediaOptions controls Storefront-backed media reads.
type ReadStoreMediaOptions struct {
	CountryCode string
	Language    string
	Kinds       []Kind
	HTTPClient  *http.Client
	MaxBytes    int64
	Concurrency int
}

// ReadStoreItemAssetOptions controls StoreBrowse-backed asset reads.
type ReadStoreItemAssetOptions struct {
	CountryCode string
	Language    string
	Kinds       []Kind
	BaseURL     string
	StripQuery  bool
	HTTPClient  *http.Client
	MaxBytes    int64
	Concurrency int
}

// Manifest is a JSON-friendly snapshot of constructed or downloaded assets.
type Manifest struct {
	URLs      []URLItem        `json:"urls,omitempty"`
	Downloads []DownloadResult `json:"downloads,omitempty"`
}
