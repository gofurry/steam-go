package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/wishlistservice"
	"github.com/GoFurry/steam-go/examples/live/internal/realtest"
)

func main() {
	cfg, err := realtest.LoadConfig()
	if err != nil {
		realtest.Fatalf("load config failed: %v", err)
	}

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	realtest.PrintProxy(cfg)

	if realtest.RequireAPIKey(cfg) {
		fmt.Println("== WishlistService.GetWishlist ==")

		wishlist, err := client.API.WishlistService.GetWishlist(
			realtest.BackgroundContext(),
			realtest.DefaultSteamID,
		)
		if err != nil {
			realtest.Fatalf("GetWishlist failed: %v", err)
		}
		fmt.Printf("items=%d\n", len(wishlist.Response.Items))

		fmt.Println("== WishlistService.GetWishlistItemCount ==")

		count, err := client.API.WishlistService.GetWishlistItemCount(
			realtest.BackgroundContext(),
			realtest.DefaultSteamID,
		)
		if err != nil {
			realtest.Fatalf("GetWishlistItemCount failed: %v", err)
		}
		fmt.Printf("count=%d\n", count.Response.Count)
	}

	if !realtest.RequireAccessToken(cfg) {
		return
	}

	fmt.Println("== WishlistService.GetWishlistItemsOnSale ==")

	trueValue := true
	resp, err := client.API.WishlistService.GetWishlistItemsOnSale(
		realtest.BackgroundContext(),
		cfg.AccessToken,
		"CN",
		&wishlistservice.GetWishlistItemsOnSaleOptions{
			IncludeAssets:                 &trueValue,
			IncludeRelease:                &trueValue,
			IncludePlatforms:              &trueValue,
			IncludeAllPurchaseOptions:     &trueValue,
			IncludeScreenshots:            &trueValue,
			IncludeTrailers:               &trueValue,
			IncludeRatings:                &trueValue,
			IncludeTagCount:               "5",
			IncludeReviews:                &trueValue,
			IncludeBasicInfo:              &trueValue,
			IncludeSupportedLanguages:     &trueValue,
			IncludeFullDescription:        &trueValue,
			IncludeIncludedItems:          &trueValue,
			IncludeAssetsWithoutOverrides: &trueValue,
			ApplyUserFilters:              &trueValue,
			IncludeLinks:                  &trueValue,
		},
	)
	if err != nil {
		realtest.Fatalf("GetWishlistItemsOnSale failed: %v", err)
	}
	fmt.Printf("items=%d total_items_on_sale=%d\n", len(resp.Response.Items), resp.Response.TotalItemsOnSale)
}
