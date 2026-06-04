package market_test

import (
	"context"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/web/market"
)

func ExampleService_GetPriceOverviewBatch() {
	client, err := steam.NewClient(steam.WithSafeDefaults())
	if err != nil {
		panic(err)
	}
	defer client.Close()

	results, err := client.Web.Market.GetPriceOverviewBatch(
		context.Background(),
		[]market.PriceOverviewBatchItem{
			{AppID: 440, MarketHashName: "Mann Co. Supply Crate Key"},
			{AppID: 730, MarketHashName: "AK-47 | Redline (Field-Tested)"},
		},
		&market.GetPriceOverviewBatchOptions{Currency: 1, MaxConcurrent: 2},
	)
	if err != nil {
		panic(err)
	}
	for _, result := range results {
		if result.Err != nil {
			continue
		}
		_ = result.Response.LowestPrice
	}
}
