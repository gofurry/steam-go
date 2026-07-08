package assets

import "github.com/gofurry/steam-go/internal/steamstatic"

const (
	// DefaultStaticCDNBaseURL is the canonical Steam static CDN base URL used by
	// the AppID-only local URL builders.
	DefaultStaticCDNBaseURL = StaticCDNSharedSteamStaticBaseURL

	// StaticCDNSharedSteamStaticBaseURL is the canonical Steam static CDN base URL.
	StaticCDNSharedSteamStaticBaseURL = steamstatic.SharedSteamStaticBaseURL

	// StaticCDNSharedAkamaiBaseURL is an Akamai-backed Steam static CDN base URL.
	StaticCDNSharedAkamaiBaseURL = steamstatic.SharedAkamaiBaseURL

	// StaticCDNSharedCloudflareBaseURL is a Cloudflare-backed Steam static CDN base URL.
	StaticCDNSharedCloudflareBaseURL = steamstatic.SharedCloudflareBaseURL

	// StaticCDNSharedFastlyBaseURL is a Fastly-backed Steam static CDN base URL.
	StaticCDNSharedFastlyBaseURL = steamstatic.SharedFastlyBaseURL

	// StaticCDNSharedSteamChinaBaishanBaseURL is a Steam China Baishan CDN base URL.
	StaticCDNSharedSteamChinaBaishanBaseURL = steamstatic.SharedSteamChinaBaishanBaseURL

	// StaticCDNSharedSteamChinaBaseURL is a Steam China static CDN base URL.
	StaticCDNSharedSteamChinaBaseURL = steamstatic.SharedSteamChinaBaseURL
)

// StaticCDNBaseURLs returns known Steam static resource CDN base URLs.
//
// The returned slice is a copy. Base URLs include a trailing slash.
func StaticCDNBaseURLs() []string {
	return steamstatic.CDNBaseURLs()
}

// StaticCDNBaseURLsWithPath returns known Steam static resource CDN base URLs
// with path appended.
//
// The returned slice is a copy. Leading path separators are normalized.
func StaticCDNBaseURLsWithPath(path string) []string {
	return steamstatic.CDNBaseURLsWithPath(path)
}

// StaticStoreItemAssetBaseURLs returns known Steam static CDN base URLs for
// Store item asset paths.
//
// The returned slice is a copy. Values include the trailing "store_item_assets/" path.
func StaticStoreItemAssetBaseURLs() []string {
	return steamstatic.StoreItemAssetBaseURLs()
}
