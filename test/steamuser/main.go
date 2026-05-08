package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/test/internal/realtest"
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
}
