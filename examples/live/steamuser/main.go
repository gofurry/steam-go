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
	if !realtest.RequireAPIKey(cfg) {
		return
	}

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	realtest.PrintProxy(cfg)
	fmt.Println("== SteamUser.GetPlayerSummaries ==")

	resp, err := client.API.SteamUser.GetPlayerSummaries(realtest.BackgroundContext(), []string{realtest.DefaultSteamID})
	if err != nil {
		realtest.Fatalf("GetPlayerSummaries failed: %v", err)
	}

	fmt.Printf("players=%d\n", len(resp.Response.Players))
	for _, player := range resp.Response.Players {
		fmt.Printf("steamid=%s persona=%s\n", player.SteamID, player.PersonaName)
	}

	fmt.Println("== SteamUser.GetFriendList ==")

	friendList, err := client.API.SteamUser.GetFriendList(realtest.BackgroundContext(), realtest.DefaultSteamID, nil)
	if err != nil {
		realtest.Fatalf("GetFriendList failed: %v", err)
	}
	fmt.Printf("friends=%d\n", len(friendList.FriendsList.Friends))
	for _, friend := range friendList.FriendsList.Friends {
		fmt.Printf("friend_steamid=%s relationship=%s\n", friend.SteamID, friend.Relationship)
	}

	fmt.Println("== SteamUser.GetPlayerBans ==")

	bans, err := client.API.SteamUser.GetPlayerBans(realtest.BackgroundContext(), []string{realtest.DefaultSteamID})
	if err != nil {
		realtest.Fatalf("GetPlayerBans failed: %v", err)
	}
	fmt.Printf("players=%d\n", len(bans.Players))
	for _, player := range bans.Players {
		fmt.Printf("steamid=%s vac_banned=%t game_bans=%d\n", player.SteamID, player.VACBanned, player.NumberOfGameBans)
	}

	fmt.Println("== SteamUser.GetUserGroupList ==")

	groupList, err := client.API.SteamUser.GetUserGroupList(realtest.BackgroundContext(), realtest.DefaultSteamID)
	if err != nil {
		realtest.Fatalf("GetUserGroupList failed: %v", err)
	}
	fmt.Printf("groups=%d success=%t\n", len(groupList.Response.Groups), groupList.Response.Success)
	for _, group := range groupList.Response.Groups {
		fmt.Printf("gid=%s\n", group.GID)
	}
}
