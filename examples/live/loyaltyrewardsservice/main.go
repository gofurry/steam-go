package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/loyaltyrewardsservice"
	"github.com/GoFurry/steam-go/examples/live/internal/realtest"
)

func main() {
	cfg, err := realtest.LoadConfig()
	if err != nil {
		realtest.Fatalf("load config failed: %v", err)
	}
	if !realtest.RequireAccessToken(cfg) {
		return
	}

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	ctx := realtest.BackgroundContext()
	realtest.PrintProxy(cfg)

	fmt.Println("== LoyaltyRewardsService.GetEquippedProfileItems ==")
	profileItemsResp, err := client.API.LoyaltyRewardsService.GetEquippedProfileItems(
		ctx,
		realtest.DefaultSteamID,
		&loyaltyrewardsservice.GetEquippedProfileItemsOptions{Language: "zh"},
	)
	if err != nil {
		realtest.Fatalf("GetEquippedProfileItems failed: %v", err)
	}
	fmt.Printf("active=%d inactive=%d\n", len(profileItemsResp.Response.ActiveDefinitions), len(profileItemsResp.Response.InactiveDefinitions))

	fmt.Println("\n== LoyaltyRewardsService.GetReactionsSummaryForUser ==")
	reactionsResp, err := client.API.LoyaltyRewardsService.GetReactionsSummaryForUser(ctx, realtest.DefaultSteamID)
	if err != nil {
		realtest.Fatalf("GetReactionsSummaryForUser failed: %v", err)
	}
	fmt.Printf("total_given=%d total_received=%d\n", reactionsResp.Response.TotalGiven, reactionsResp.Response.TotalReceived)

	fmt.Println("\n== LoyaltyRewardsService.GetSummary ==")
	summaryResp, err := client.API.LoyaltyRewardsService.GetSummary(ctx, realtest.DefaultSteamID)
	if err != nil {
		realtest.Fatalf("GetSummary failed: %v", err)
	}
	fmt.Printf("points=%s earned=%s spent=%s\n", summaryResp.Response.Summary.Points, summaryResp.Response.Summary.PointsEarned, summaryResp.Response.Summary.PointsSpent)
}
