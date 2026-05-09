package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/userstorevisitservice"
	"github.com/GoFurry/steam-go/test/internal/realtest"
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

	if realtest.RequireAccessToken(cfg) {
		fmt.Println("== UserStoreVisitService.GetFrequentlyVisitedPages ==")

		visits, err := client.API.UserStoreVisitService.GetFrequentlyVisitedPages(
			realtest.BackgroundContext(),
			cfg.AccessToken,
		)
		if err != nil {
			realtest.Fatalf("GetFrequentlyVisitedPages failed: %v", err)
		}
		fmt.Printf("recent_apps=%d frequent_hubs=%d\n",
			len(visits.Response.VisitData.RecentApps),
			len(visits.Response.FrequentHubs),
		)
	}

	if !realtest.RequireAPIKey(cfg) {
		return
	}

	fmt.Println("== UserStoreVisitService.GetMostVisitedItemsOnStore ==")

	trueValue := true
	resp, err := client.API.UserStoreVisitService.GetMostVisitedItemsOnStore(
		realtest.BackgroundContext(),
		"CN",
		&userstorevisitservice.GetMostVisitedItemsOnStoreOptions{
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
		realtest.Fatalf("GetMostVisitedItemsOnStore failed: %v", err)
	}
	fmt.Printf("item_ids=%d items=%d\n", len(resp.Response.ItemIDs), len(resp.Response.Items))
}
