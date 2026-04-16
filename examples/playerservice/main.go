package main

import (
	"context"
	"fmt"
	"os"
	"time"

	steam "github.com/GoFurry/steam-go"
	"github.com/GoFurry/steam-go/api/playerservice"
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

	resp, err := client.API.PlayerService.GetOwnedGames(
		context.Background(),
		"76561198370695025",
		&playerservice.GetOwnedGamesOptions{IncludePlayedFreeGames: true},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("games: %d\n", resp.Response.GameCount)
	for _, game := range resp.Response.Games {
		fmt.Printf("%d: %s\n", game.AppID, game.Name)
	}
}
