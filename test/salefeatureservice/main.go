package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/salefeatureservice"
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

	ctx := realtest.BackgroundContext()
	realtest.PrintProxy(cfg)

	returnPrivate := false
	fmt.Println("== SaleFeatureService.GetFriendsSharedYearInReview ==")
	friendSharesResp, err := client.API.SaleFeatureService.GetFriendsSharedYearInReview(
		ctx,
		realtest.DefaultSteamID,
		2025,
		&salefeatureservice.GetFriendsSharedYearInReviewOptions{ReturnPrivate: &returnPrivate},
	)
	if err != nil {
		realtest.Fatalf("GetFriendsSharedYearInReview failed: %v", err)
	}
	fmt.Printf("friend_shares=%d\n", len(friendSharesResp.Response.FriendShares))

	fmt.Println("\n== SaleFeatureService.GetUserYearInReview ==")
	yearInReviewResp, err := client.API.SaleFeatureService.GetUserYearInReview(
		ctx,
		realtest.DefaultSteamID,
		2025,
	)
	if err != nil {
		realtest.Fatalf("GetUserYearInReview failed: %v", err)
	}
	fmt.Printf("account_id=%d substantial=%t total_playtime_seconds=%d games=%d\n",
		yearInReviewResp.Response.Stats.AccountID,
		yearInReviewResp.Response.Stats.Substantial,
		yearInReviewResp.Response.Stats.PlaytimeStats.TotalStats.TotalPlaytimeSeconds,
		len(yearInReviewResp.Response.Stats.PlaytimeStats.Games),
	)

	if !realtest.RequireAccessToken(cfg) {
		return
	}

	totalOnly := false
	fmt.Println("\n== SaleFeatureService.GetUserYearAchievements ==")
	yearAchievementsResp, err := client.API.SaleFeatureService.GetUserYearAchievements(
		ctx,
		cfg.AccessToken,
		&salefeatureservice.GetUserYearAchievementsOptions{
			SteamID:   realtest.DefaultSteamID,
			Year:      2022,
			AppIDs:    []uint32{realtest.DefaultAppID},
			TotalOnly: &totalOnly,
		},
	)
	if err != nil {
		realtest.Fatalf("GetUserYearAchievements failed: %v", err)
	}
	fmt.Printf("total_achievements=%d total_rare_achievements=%d games=%d\n",
		yearAchievementsResp.Response.TotalAchievements,
		yearAchievementsResp.Response.TotalRareAchievements,
		len(yearAchievementsResp.Response.GameAchievements),
	)
}
