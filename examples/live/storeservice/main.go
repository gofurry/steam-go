package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/storeservice"
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
		fmt.Println("== StoreService.GetAppList ==")

		trueValue := true
		lastAppID := realtest.DefaultAppID
		maxResults := uint32(10)
		appList, err := client.API.StoreService.GetAppList(realtest.BackgroundContext(), &storeservice.GetAppListOptions{
			IncludeGames:    &trueValue,
			IncludeDLC:      &trueValue,
			IncludeSoftware: &trueValue,
			IncludeVideos:   &trueValue,
			IncludeHardware: &trueValue,
			LastAppID:       &lastAppID,
			MaxResults:      &maxResults,
		})
		if err != nil {
			realtest.Fatalf("GetAppList failed: %v", err)
		}
		fmt.Printf("apps=%d have_more=%t last_appid=%d\n", len(appList.Response.Apps), appList.Response.HaveMoreResults, appList.Response.LastAppID)

		fmt.Println("== StoreService.GetGamesFollowed ==")

		followed, err := client.API.StoreService.GetGamesFollowed(realtest.BackgroundContext(), realtest.DefaultSteamID)
		if err != nil {
			realtest.Fatalf("GetGamesFollowed failed: %v", err)
		}
		fmt.Printf("followed_games=%d\n", len(followed.Response.AppIDs))

		fmt.Println("== StoreService.GetGamesFollowedCount ==")

		count, err := client.API.StoreService.GetGamesFollowedCount(realtest.BackgroundContext(), realtest.DefaultSteamID)
		if err != nil {
			realtest.Fatalf("GetGamesFollowedCount failed: %v", err)
		}
		fmt.Printf("followed_game_count=%d\n", count.Response.FollowedGameCount)

		fmt.Println("== StoreService.GetMostPopularTags ==")

		tags, err := client.API.StoreService.GetMostPopularTags(realtest.BackgroundContext())
		if err != nil {
			realtest.Fatalf("GetMostPopularTags failed: %v", err)
		}
		fmt.Printf("tags=%d\n", len(tags.Response.Tags))
		for i, tag := range tags.Response.Tags {
			if i >= 10 {
				break
			}
			fmt.Printf("[%d] tagid=%d name=%s\n", i+1, tag.TagID, tag.Name)
		}
	}

	if realtest.RequireAccessToken(cfg) {
		fmt.Println("== StoreService.GetUserGameInterestState ==")

		interestState, err := client.API.StoreService.GetUserGameInterestState(
			realtest.BackgroundContext(),
			cfg.AccessToken,
			realtest.DefaultAppID,
			nil,
		)
		if err != nil {
			realtest.Fatalf("GetUserGameInterestState failed: %v", err)
		}
		fmt.Printf("owned=%t following=%t queues=%d\n",
			interestState.Response.Owned,
			interestState.Response.Following,
			len(interestState.Response.Queues),
		)
	}
}
