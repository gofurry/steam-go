package urlbuilder

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	StaticBaseURL         = "https://shared.steamstatic.com"
	CommunityImageBaseURL = "https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps"
)

func StoreAsset(appID uint32, filename string) string {
	return StaticBaseURL + "/store_item_assets/steam/apps/" + strconv.FormatUint(uint64(appID), 10) + "/" + filename
}

func Header(appID uint32) string {
	return StoreAsset(appID, "header.jpg")
}

func HeaderLocalized(appID uint32, language string) string {
	language = strings.TrimSpace(language)
	if !SafePathToken(language) {
		return ""
	}
	return StoreAsset(appID, fmt.Sprintf("header_%s.jpg", language))
}

func CapsuleSmall(appID uint32) string {
	return StoreAsset(appID, "capsule_231x87.jpg")
}

func CapsuleMain(appID uint32) string {
	return StoreAsset(appID, "capsule_616x353.jpg")
}

func LibraryCapsule(appID uint32) string {
	return StoreAsset(appID, "library_600x900.jpg")
}

func LibraryCapsule2x(appID uint32) string {
	return StoreAsset(appID, "library_600x900_2x.jpg")
}

func LibraryHero(appID uint32) string {
	return StoreAsset(appID, "library_hero.jpg")
}

func LibraryLogo(appID uint32) string {
	return StoreAsset(appID, "logo.png")
}

func LibraryLogo2x(appID uint32) string {
	return StoreAsset(appID, "logo_2x.png")
}

func CommunityImage(appID uint32, hash, ext string) string {
	hash = strings.TrimSpace(hash)
	if !SafeHash(hash) {
		return ""
	}
	return CommunityImageBaseURL + "/" + strconv.FormatUint(uint64(appID), 10) + "/" + hash + "." + ext
}

func SafePathToken(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func SafeHash(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}
