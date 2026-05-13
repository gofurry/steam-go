package main

import (
	"fmt"

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

	if !realtest.RequireAccessToken(cfg) {
		return
	}

	fmt.Println("== UserReviewsService.GetFriendsRecommendedApp ==")

	resp, err := client.API.UserReviewsService.GetFriendsRecommendedApp(
		realtest.BackgroundContext(),
		cfg.AccessToken,
		realtest.DefaultAppID,
	)
	if err != nil {
		realtest.Fatalf("GetFriendsRecommendedApp failed: %v", err)
	}
	fmt.Printf("recommended_accounts=%d\n", len(resp.Response.AccountIDsRecommended))
	for i, accountID := range resp.Response.AccountIDsRecommended {
		if i >= 10 {
			break
		}
		fmt.Printf("[%d] %d\n", i+1, accountID)
	}
}
