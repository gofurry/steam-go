package storefront_test

import (
	"context"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/web/storefront"
)

func ExampleService_ListAppReviews() {
	client, err := steam.NewClient(steam.WithSafeDefaults())
	if err != nil {
		panic(err)
	}
	defer client.Close()

	err = client.Web.Storefront.ListAppReviews(
		context.Background(),
		550,
		&storefront.ListAppReviewsOptions{MaxPages: 2},
		func(page storefront.AppReviewsPage) error {
			for _, review := range page.Reviews {
				_ = review.RecommendationID
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
}

func ExampleService_GetAppDetailsBatch() {
	client, err := steam.NewClient(steam.WithSafeDefaults())
	if err != nil {
		panic(err)
	}
	defer client.Close()

	results, err := client.Web.Storefront.GetAppDetailsBatch(
		context.Background(),
		[]uint32{550, 440},
		&storefront.GetAppDetailsBatchOptions{MaxConcurrent: 2},
	)
	if err != nil {
		panic(err)
	}
	for _, result := range results {
		if result.Err != nil {
			continue
		}
		_ = result.Response
	}
}
