package assets

import (
	"reflect"
	"testing"
)

func TestResourceSpecificURLsPreserveOrder(t *testing.T) {
	got := HeaderURLs(550, 107100)
	want := []string{
		"https://shared.steamstatic.com/store_item_assets/steam/apps/550/header.jpg",
		"https://shared.steamstatic.com/store_item_assets/steam/apps/107100/header.jpg",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HeaderURLs() = %#v, want %#v", got, want)
	}

	got = LibraryLogo2xURLs(550, 107100)
	want = []string{
		"https://shared.steamstatic.com/store_item_assets/steam/apps/550/logo_2x.png",
		"https://shared.steamstatic.com/store_item_assets/steam/apps/107100/logo_2x.png",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LibraryLogo2xURLs() = %#v, want %#v", got, want)
	}
}

func TestLocalizedHeaderURLs(t *testing.T) {
	got := HeaderLocalizedURLs("schinese", 550)
	want := []string{"https://shared.steamstatic.com/store_item_assets/steam/apps/550/header_schinese.jpg"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HeaderLocalizedURLs() = %#v, want %#v", got, want)
	}

	got = HeaderLocalizedURLs("../bad", 550, 107100)
	want = []string{"", ""}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HeaderLocalizedURLs() with unsafe language = %#v, want %#v", got, want)
	}
}

func TestURLsByKind(t *testing.T) {
	got := URLs(KindCapsuleMain, 550, 107100)
	want := []string{
		"https://shared.steamstatic.com/store_item_assets/steam/apps/550/capsule_616x353.jpg",
		"https://shared.steamstatic.com/store_item_assets/steam/apps/107100/capsule_616x353.jpg",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("URLs() = %#v, want %#v", got, want)
	}

	got = LocalizedURLs(KindHeaderLocalized, "japanese", 550)
	want = []string{"https://shared.steamstatic.com/store_item_assets/steam/apps/550/header_japanese.jpg"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LocalizedURLs() = %#v, want %#v", got, want)
	}

	got = URLs(Kind("unknown"), 550, 107100)
	want = []string{"", ""}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("URLs() with unknown kind = %#v, want %#v", got, want)
	}
}

func TestAll(t *testing.T) {
	got := AllWithLanguage("schinese", 550)
	if len(got) != 1 {
		t.Fatalf("AllWithLanguage() returned %d items, want 1", len(got))
	}
	if got[0].AppID != 550 {
		t.Fatalf("AppID = %d, want 550", got[0].AppID)
	}
	if got[0].Header != "https://shared.steamstatic.com/store_item_assets/steam/apps/550/header.jpg" {
		t.Fatalf("Header = %q", got[0].Header)
	}
	if got[0].HeaderLocalized != "https://shared.steamstatic.com/store_item_assets/steam/apps/550/header_schinese.jpg" {
		t.Fatalf("HeaderLocalized = %q", got[0].HeaderLocalized)
	}
	if got[0].LibraryHero != "https://shared.steamstatic.com/store_item_assets/steam/apps/550/library_hero.jpg" {
		t.Fatalf("LibraryHero = %q", got[0].LibraryHero)
	}
}

func TestHashURLs(t *testing.T) {
	refs := []HashRef{{AppID: 550, Hash: "abcdef0123456789"}}

	got := CommunityIconURLs(refs...)
	want := []string{"https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/550/abcdef0123456789.jpg"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CommunityIconURLs() = %#v, want %#v", got, want)
	}

	got = ClientIconURLs(refs...)
	want = []string{"https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/550/abcdef0123456789.ico"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ClientIconURLs() = %#v, want %#v", got, want)
	}

	got = ClientIconURLs(HashRef{AppID: 550, Hash: "../bad"})
	want = []string{""}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ClientIconURLs() with unsafe hash = %#v, want %#v", got, want)
	}
}

func TestKindGroupsAndList(t *testing.T) {
	if got, want := StoreKinds(), []Kind{KindHeader, KindCapsuleSmall, KindCapsuleMain}; !reflect.DeepEqual(got, want) {
		t.Fatalf("StoreKinds() = %#v, want %#v", got, want)
	}
	if got, want := LibraryKinds(), []Kind{KindLibraryCapsule, KindLibraryCapsule2x, KindLibraryHero, KindLibraryLogo, KindLibraryLogo2x}; !reflect.DeepEqual(got, want) {
		t.Fatalf("LibraryKinds() = %#v, want %#v", got, want)
	}
	if got := StoreMediaKinds(); len(got) == 0 || got[0] != KindScreenshotThumbnail {
		t.Fatalf("StoreMediaKinds() = %#v", got)
	}

	items := ListKindsWithLanguage("schinese", []Kind{KindHeaderLocalized, KindLibraryHero}, 550)
	want := []URLItem{
		{AppID: 550, Kind: KindHeaderLocalized, URL: "https://shared.steamstatic.com/store_item_assets/steam/apps/550/header_schinese.jpg"},
		{AppID: 550, Kind: KindLibraryHero, URL: "https://shared.steamstatic.com/store_item_assets/steam/apps/550/library_hero.jpg"},
	}
	if !reflect.DeepEqual(items, want) {
		t.Fatalf("ListKindsWithLanguage() = %#v, want %#v", items, want)
	}

	if got := CommunityIconURL(550, "abcdef0123456789"); got != "https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/550/abcdef0123456789.jpg" {
		t.Fatalf("CommunityIconURL() = %q", got)
	}
	if got := ClientIconURL(550, "abcdef0123456789"); got != "https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/550/abcdef0123456789.ico" {
		t.Fatalf("ClientIconURL() = %q", got)
	}
}
