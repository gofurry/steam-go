package main

import (
	"fmt"
	"time"

	"github.com/gofurry/steam-go/web/storefront"
)

func main() {
	releaseDates := []*storefront.StoreReleaseDate{
		{Date: "16 Nov, 2009"},
		{ComingSoon: true, Date: "Q3 2027"},
		{ComingSoon: true, Date: "Coming Soon"},
		{ComingSoon: true, Date: "Late 2027"},
	}
	for _, releaseDate := range releaseDates {
		normalized := storefront.NormalizeReleaseDate(releaseDate)
		fmt.Printf(
			"release raw=%q coming_soon=%t precision=%s range=%s..%s\n",
			normalized.RawText,
			normalized.ComingSoon,
			normalized.Precision,
			formatDate(normalized.RangeStart),
			formatDate(normalized.RangeEnd),
		)
	}

	rawLanguages := `English<strong>*</strong>, Japanese, Korean<strong>*</strong>, Example Future Language<br><strong>*</strong>languages with full audio support`
	for _, language := range storefront.ParseSupportedLanguages(rawLanguages) {
		fmt.Printf(
			"language name=%q code=%q tier=%s known=%t full_audio=%s\n",
			language.SteamName,
			language.Code,
			language.Tier,
			language.Known,
			formatOptionalBool(language.FullAudio),
		)
	}

	definition, ok := storefront.LookupLanguage("schinese")
	fmt.Printf("lookup schinese ok=%t canonical=%s web=%s\n", ok, definition.Code, definition.SteamWebCode)
}

func formatDate(value *time.Time) string {
	if value == nil {
		return "-"
	}
	return value.Format("2006-01-02")
}

func formatOptionalBool(value *bool) string {
	if value == nil {
		return "unknown"
	}
	return fmt.Sprint(*value)
}
