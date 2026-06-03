package assets_test

import (
	"fmt"

	"github.com/gofurry/steam-go/addons/assets"
)

func Example() {
	headers := assets.HeaderURLs(440)
	heroes := assets.URLs(assets.KindLibraryHero, 440)

	fmt.Println(headers[0])
	fmt.Println(heroes[0])

	// Output:
	// https://shared.steamstatic.com/store_item_assets/steam/apps/440/header.jpg
	// https://shared.steamstatic.com/store_item_assets/steam/apps/440/library_hero.jpg
}
