package main

import (
	"context"
	"fmt"
	"os"
	"time"

	steam "github.com/GoFurry/steam-go"
)

func main() {
	apiKey := os.Getenv("STEAM_API_KEY")
	if apiKey == "" {
		fmt.Println("STEAM_API_KEY is required")
		return
	}

	client, err := steam.NewClient(
		steam.WithAPIKey(apiKey),
		steam.WithTimeout(10*time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	resp, err := client.API.SteamUser.GetPlayerSummaries(
		context.Background(),
		[]string{"76561198370695025"},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("players: %d\n", len(resp.Response.Players))
	for _, player := range resp.Response.Players {
		fmt.Printf("%s: %s\n", player.SteamID, player.PersonaName)
	}
}
