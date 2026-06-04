package community_test

import (
	"context"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/web/community"
)

func ExampleService_ListInventory() {
	client, err := steam.NewClient(steam.WithDefaultCookieJar(), steam.WithSafeDefaults())
	if err != nil {
		panic(err)
	}
	defer client.Close()

	err = client.Web.Community.ListInventory(
		context.Background(),
		"76561198370695025",
		730,
		"2",
		&community.ListInventoryOptions{MaxPages: 2},
		func(page community.InventoryPage) error {
			for _, asset := range page.Assets {
				_ = asset.AssetID
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
}
