package assets

import "github.com/gofurry/steam-go/addons/assets/internal/urlbuilder"

// StoreKinds returns Store asset kinds that do not require external hashes.
func StoreKinds() []Kind {
	return []Kind{KindHeader, KindCapsuleSmall, KindCapsuleMain}
}

// StoreKindsWithLocalized returns Store asset kinds including localized headers.
func StoreKindsWithLocalized() []Kind {
	return append([]Kind{KindHeaderLocalized}, StoreKinds()...)
}

// LibraryKinds returns Library asset kinds.
func LibraryKinds() []Kind {
	return []Kind{
		KindLibraryCapsule,
		KindLibraryCapsule2x,
		KindLibraryHero,
		KindLibraryLogo,
		KindLibraryLogo2x,
	}
}

// StoreMediaKinds returns Storefront-backed media kinds.
func StoreMediaKinds() []Kind {
	kinds := StoreScreenshotKinds()
	kinds = append(kinds, StoreMovieKinds()...)
	kinds = append(kinds, StoreBackgroundKinds()...)
	return kinds
}

// StoreBackgroundKinds returns Storefront-backed background image kinds.
func StoreBackgroundKinds() []Kind {
	return []Kind{KindStoreBackground, KindStoreBackgroundRaw}
}

// StoreScreenshotKinds returns Storefront-backed screenshot image kinds.
func StoreScreenshotKinds() []Kind {
	return []Kind{KindScreenshotThumbnail, KindScreenshotFull}
}

// StoreMovieKinds returns Storefront-backed movie image and playlist kinds.
func StoreMovieKinds() []Kind {
	return []Kind{
		KindMovieThumbnail,
		KindMovieWebM480,
		KindMovieWebMMax,
		KindMovieMP4480,
		KindMovieMP4Max,
		KindMovieDASHAV1,
		KindMovieDASHH264,
		KindMovieHLSH264,
	}
}

// DefaultKinds returns all standard Store and Library kinds that do not need language or hashes.
func DefaultKinds() []Kind {
	kinds := StoreKinds()
	kinds = append(kinds, LibraryKinds()...)
	return kinds
}

// AllKinds returns all locally constructible Store and Library kinds, including localized headers.
func AllKinds() []Kind {
	kinds := StoreKindsWithLocalized()
	kinds = append(kinds, LibraryKinds()...)
	return kinds
}

// HeaderURLs returns Steam Store header image URLs in the same order as appIDs.
func HeaderURLs(appIDs ...uint32) []string {
	return urlsFor(appIDs, urlbuilder.Header)
}

// HeaderLocalizedURLs returns localized Steam Store header image URLs in the same order as appIDs.
//
// If language is not a safe path token, each returned URL is an empty string.
func HeaderLocalizedURLs(language string, appIDs ...uint32) []string {
	return urlsFor(appIDs, func(appID uint32) string {
		return urlbuilder.HeaderLocalized(appID, language)
	})
}

// CapsuleSmallURLs returns Steam Store small capsule image URLs in the same order as appIDs.
func CapsuleSmallURLs(appIDs ...uint32) []string {
	return urlsFor(appIDs, urlbuilder.CapsuleSmall)
}

// CapsuleMainURLs returns Steam Store main capsule image URLs in the same order as appIDs.
func CapsuleMainURLs(appIDs ...uint32) []string {
	return urlsFor(appIDs, urlbuilder.CapsuleMain)
}

// LibraryCapsuleURLs returns Steam Library capsule image URLs in the same order as appIDs.
func LibraryCapsuleURLs(appIDs ...uint32) []string {
	return urlsFor(appIDs, urlbuilder.LibraryCapsule)
}

// LibraryCapsule2xURLs returns Steam Library 2x capsule image URLs in the same order as appIDs.
func LibraryCapsule2xURLs(appIDs ...uint32) []string {
	return urlsFor(appIDs, urlbuilder.LibraryCapsule2x)
}

// LibraryHeroURLs returns Steam Library hero image URLs in the same order as appIDs.
func LibraryHeroURLs(appIDs ...uint32) []string {
	return urlsFor(appIDs, urlbuilder.LibraryHero)
}

// LibraryLogoURLs returns Steam Library logo image URLs in the same order as appIDs.
func LibraryLogoURLs(appIDs ...uint32) []string {
	return urlsFor(appIDs, urlbuilder.LibraryLogo)
}

// LibraryLogo2xURLs returns Steam Library 2x logo image URLs in the same order as appIDs.
func LibraryLogo2xURLs(appIDs ...uint32) []string {
	return urlsFor(appIDs, urlbuilder.LibraryLogo2x)
}

// URLs returns URLs for one static asset kind in the same order as appIDs.
//
// For KindHeaderLocalized, use LocalizedURLs so the language can be supplied.
// Unknown kinds return empty strings while preserving result length and order.
func URLs(kind Kind, appIDs ...uint32) []string {
	switch kind {
	case KindHeader:
		return HeaderURLs(appIDs...)
	case KindCapsuleSmall:
		return CapsuleSmallURLs(appIDs...)
	case KindCapsuleMain:
		return CapsuleMainURLs(appIDs...)
	case KindLibraryCapsule:
		return LibraryCapsuleURLs(appIDs...)
	case KindLibraryCapsule2x:
		return LibraryCapsule2xURLs(appIDs...)
	case KindLibraryHero:
		return LibraryHeroURLs(appIDs...)
	case KindLibraryLogo:
		return LibraryLogoURLs(appIDs...)
	case KindLibraryLogo2x:
		return LibraryLogo2xURLs(appIDs...)
	default:
		return emptyURLs(len(appIDs))
	}
}

// LocalizedURLs returns localized URLs for one static asset kind in the same order as appIDs.
//
// The first version supports KindHeaderLocalized. Other kinds return empty strings.
func LocalizedURLs(kind Kind, language string, appIDs ...uint32) []string {
	if kind != KindHeaderLocalized {
		return emptyURLs(len(appIDs))
	}
	return HeaderLocalizedURLs(language, appIDs...)
}

// All returns all standard Store and Library static asset URLs for each AppID.
func All(appIDs ...uint32) []AppAssets {
	out := make([]AppAssets, 0, len(appIDs))
	for _, appID := range appIDs {
		out = append(out, appAssets(appID, ""))
	}
	return out
}

// AllWithLanguage returns all standard Store and Library static asset URLs for each AppID,
// including localized header URLs for the supplied language.
func AllWithLanguage(language string, appIDs ...uint32) []AppAssets {
	out := make([]AppAssets, 0, len(appIDs))
	for _, appID := range appIDs {
		out = append(out, appAssets(appID, language))
	}
	return out
}

// List returns one URL item per default Store and Library asset kind for each AppID.
func List(appIDs ...uint32) []URLItem {
	return ListKinds(DefaultKinds(), appIDs...)
}

// ListWithLanguage returns one URL item per Store and Library asset kind for each AppID,
// including localized headers for the supplied language.
func ListWithLanguage(language string, appIDs ...uint32) []URLItem {
	return ListKindsWithLanguage(language, AllKinds(), appIDs...)
}

// ListKinds returns one URL item per requested kind for each AppID.
func ListKinds(kinds []Kind, appIDs ...uint32) []URLItem {
	return ListKindsWithLanguage("", kinds, appIDs...)
}

// ListKindsWithLanguage returns one URL item per requested kind for each AppID,
// using language for localized kinds.
func ListKindsWithLanguage(language string, kinds []Kind, appIDs ...uint32) []URLItem {
	if len(kinds) == 0 {
		kinds = DefaultKinds()
	}
	out := make([]URLItem, 0, len(kinds)*len(appIDs))
	for _, appID := range appIDs {
		for _, kind := range kinds {
			rawURL := urlForKind(kind, language, appID)
			if rawURL == "" {
				continue
			}
			out = append(out, URLItem{AppID: appID, Kind: kind, URL: rawURL})
		}
	}
	return out
}

// CommunityIconURLs returns Steam community icon JPG URLs from AppID/hash pairs.
func CommunityIconURLs(refs ...HashRef) []string {
	return hashURLs(refs, "jpg")
}

// CommunityIconURL returns one Steam community icon JPG URL from an AppID/hash pair.
func CommunityIconURL(appID uint32, hash string) string {
	return CommunityIconURLs(HashRef{AppID: appID, Hash: hash})[0]
}

// CommunityLogoURLs returns Steam community logo JPG URLs from AppID/hash pairs.
func CommunityLogoURLs(refs ...HashRef) []string {
	return hashURLs(refs, "jpg")
}

// CommunityLogoURL returns one Steam community logo JPG URL from an AppID/hash pair.
func CommunityLogoURL(appID uint32, hash string) string {
	return CommunityLogoURLs(HashRef{AppID: appID, Hash: hash})[0]
}

// ClientIconURLs returns Steam client icon ICO URLs from AppID/hash pairs.
func ClientIconURLs(refs ...HashRef) []string {
	return hashURLs(refs, "ico")
}

// ClientIconURL returns one Steam client icon ICO URL from an AppID/hash pair.
func ClientIconURL(appID uint32, hash string) string {
	return ClientIconURLs(HashRef{AppID: appID, Hash: hash})[0]
}

func appAssets(appID uint32, language string) AppAssets {
	return AppAssets{
		AppID:            appID,
		Header:           urlbuilder.Header(appID),
		HeaderLocalized:  urlbuilder.HeaderLocalized(appID, language),
		CapsuleSmall:     urlbuilder.CapsuleSmall(appID),
		CapsuleMain:      urlbuilder.CapsuleMain(appID),
		LibraryCapsule:   urlbuilder.LibraryCapsule(appID),
		LibraryCapsule2x: urlbuilder.LibraryCapsule2x(appID),
		LibraryHero:      urlbuilder.LibraryHero(appID),
		LibraryLogo:      urlbuilder.LibraryLogo(appID),
		LibraryLogo2x:    urlbuilder.LibraryLogo2x(appID),
	}
}

func urlsFor(appIDs []uint32, build func(uint32) string) []string {
	out := make([]string, 0, len(appIDs))
	for _, appID := range appIDs {
		out = append(out, build(appID))
	}
	return out
}

func urlForKind(kind Kind, language string, appID uint32) string {
	switch kind {
	case KindHeaderLocalized:
		return HeaderLocalizedURLs(language, appID)[0]
	case KindHeader, KindCapsuleSmall, KindCapsuleMain, KindLibraryCapsule, KindLibraryCapsule2x, KindLibraryHero, KindLibraryLogo, KindLibraryLogo2x:
		return URLs(kind, appID)[0]
	default:
		return ""
	}
}

func hashURLs(refs []HashRef, ext string) []string {
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		out = append(out, urlbuilder.CommunityImage(ref.AppID, ref.Hash, ext))
	}
	return out
}

func emptyURLs(n int) []string {
	out := make([]string, n)
	return out
}
