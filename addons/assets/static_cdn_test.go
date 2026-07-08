package assets

import (
	"reflect"
	"testing"
)

func TestStaticCDNBaseURLs(t *testing.T) {
	want := []string{
		"https://shared.steamstatic.com/",
		"https://shared.akamai.steamstatic.com/",
		"https://shared.cloudflare.steamstatic.com/",
		"https://shared.fastly.steamstatic.com/",
		"https://shared.st.dl.eccdnx.com/",
		"https://shared.cdn.steamchina.queniuam.com/",
	}

	got := StaticCDNBaseURLs()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("StaticCDNBaseURLs() = %#v, want %#v", got, want)
	}

	got[0] = "https://example.com/"
	if StaticCDNBaseURLs()[0] != StaticCDNSharedSteamStaticBaseURL {
		t.Fatal("StaticCDNBaseURLs returned shared backing storage")
	}
}

func TestStaticStoreItemAssetBaseURLs(t *testing.T) {
	want := []string{
		"https://shared.steamstatic.com/store_item_assets/",
		"https://shared.akamai.steamstatic.com/store_item_assets/",
		"https://shared.cloudflare.steamstatic.com/store_item_assets/",
		"https://shared.fastly.steamstatic.com/store_item_assets/",
		"https://shared.st.dl.eccdnx.com/store_item_assets/",
		"https://shared.cdn.steamchina.queniuam.com/store_item_assets/",
	}

	if got := StaticStoreItemAssetBaseURLs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("StaticStoreItemAssetBaseURLs() = %#v, want %#v", got, want)
	}
	if got := StaticCDNBaseURLsWithPath("/store_item_assets/"); !reflect.DeepEqual(got, want) {
		t.Fatalf("StaticCDNBaseURLsWithPath() = %#v, want %#v", got, want)
	}
}
