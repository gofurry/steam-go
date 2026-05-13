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

	if realtest.RequireAPIKey(cfg) {
		fmt.Println("== SteamUserOAuth.GetUserSummaries ==")

		resp, err := client.API.SteamUserOAuth.GetUserSummaries(
			realtest.BackgroundContext(),
			[]string{realtest.DefaultSteamID, "76561198856448829"},
		)
		if err != nil {
			realtest.Fatalf("GetUserSummaries failed: %v", err)
		}
		fmt.Printf("players=%d\n", len(resp.Players))
		for _, player := range resp.Players {
			fmt.Printf("steamid=%s persona=%s\n", player.SteamID, player.PersonaName)
		}
	}

	if realtest.RequireAccessToken(cfg) {
		fmt.Println("== SteamUserOAuth.GetFriendList ==")

		friendList, err := client.API.SteamUserOAuth.GetFriendList(realtest.BackgroundContext(), cfg.AccessToken)
		if err != nil {
			realtest.Fatalf("GetFriendList failed: %v", err)
		}
		fmt.Printf("friends=%d\n", len(friendList.Friends))
		for i, friend := range friendList.Friends {
			if i >= 20 {
				break
			}
			fmt.Printf("[%d] steamid=%s relationship=%s\n", i+1, friend.SteamID, friend.Relationship)
		}
	}
}
