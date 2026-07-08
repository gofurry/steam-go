package steamstatic

import "strings"

const (
	HostSuffix = "steamstatic.com"

	SharedSteamStaticHost       = "shared.steamstatic.com"
	SharedAkamaiHost            = "shared.akamai.steamstatic.com"
	SharedCloudflareHost        = "shared.cloudflare.steamstatic.com"
	SharedFastlyHost            = "shared.fastly.steamstatic.com"
	SharedSteamChinaBaishanHost = "shared.st.dl.eccdnx.com"
	SharedSteamChinaHost        = "shared.cdn.steamchina.queniuam.com"

	SharedSteamStaticBaseURL       = "https://" + SharedSteamStaticHost + "/"
	SharedAkamaiBaseURL            = "https://" + SharedAkamaiHost + "/"
	SharedCloudflareBaseURL        = "https://" + SharedCloudflareHost + "/"
	SharedFastlyBaseURL            = "https://" + SharedFastlyHost + "/"
	SharedSteamChinaBaishanBaseURL = "https://" + SharedSteamChinaBaishanHost + "/"
	SharedSteamChinaBaseURL        = "https://" + SharedSteamChinaHost + "/"

	StoreItemAssetsPath = "store_item_assets/"
)

var cdnBaseURLs = []string{
	SharedSteamStaticBaseURL,
	SharedAkamaiBaseURL,
	SharedCloudflareBaseURL,
	SharedFastlyBaseURL,
	SharedSteamChinaBaishanBaseURL,
	SharedSteamChinaBaseURL,
}

var cdnHosts = []string{
	SharedSteamStaticHost,
	SharedAkamaiHost,
	SharedCloudflareHost,
	SharedFastlyHost,
	SharedSteamChinaBaishanHost,
	SharedSteamChinaHost,
}

var nonSteamStaticCDNHosts = []string{
	SharedSteamChinaBaishanHost,
	SharedSteamChinaHost,
}

func CDNBaseURLs() []string {
	return cloneStrings(cdnBaseURLs)
}

func CDNHosts() []string {
	return cloneStrings(cdnHosts)
}

func NonSteamStaticCDNHosts() []string {
	return cloneStrings(nonSteamStaticCDNHosts)
}

func StoreItemAssetBaseURLs() []string {
	return CDNBaseURLsWithPath(StoreItemAssetsPath)
}

func CDNBaseURLsWithPath(path string) []string {
	out := make([]string, 0, len(cdnBaseURLs))
	for _, baseURL := range cdnBaseURLs {
		out = append(out, Join(baseURL, path))
	}
	return out
}

func Join(baseURL, path string) string {
	baseURL = strings.TrimSpace(baseURL)
	path = strings.TrimSpace(path)
	if baseURL == "" {
		return ""
	}
	if path == "" {
		return strings.TrimRight(baseURL, "/") + "/"
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
}

func cloneStrings(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	out := make([]string, len(src))
	copy(out, src)
	return out
}
