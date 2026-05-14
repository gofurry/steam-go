package main

import (
	"fmt"

	"github.com/gofurry/steam-go/examples/live/internal/realtest"
	"github.com/gofurry/steam-go/web/storefront"
)

func main() {
	cfg, err := realtest.LoadConfig()
	if err != nil {
		realtest.Fatalf("load config: %v", err)
	}
	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("new client: %v", err)
	}
	defer client.Close()

	ctx := realtest.BackgroundContext()
	realtest.PrintProxy(cfg)

	appDetails, err := client.Web.Storefront.GetAppDetails(ctx, 550, nil)
	if err != nil {
		realtest.Fatalf("GetAppDetails: %v", err)
	}
	fmt.Printf("appdetails ok=%v\n", appDetails["550"].Success)

	packageDetails, err := client.Web.Storefront.GetPackageDetails(ctx, 469, nil)
	if err != nil {
		realtest.Fatalf("GetPackageDetails: %v", err)
	}
	fmt.Printf("packagedetails ok=%v\n", packageDetails["469"].Success)

	reviews, err := client.Web.Storefront.GetAppReviews(ctx, 550, &storefront.GetAppReviewsOptions{
		NumPerPage: 1,
	})
	if err != nil {
		realtest.Fatalf("GetAppReviews: %v", err)
	}
	fmt.Printf("reviews total=%d cursor=%s\n", reviews.QuerySummary.TotalReviews, reviews.Cursor)
}
