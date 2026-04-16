package main

import (
	"context"
	"fmt"
	"os"
	"time"

	steam "github.com/GoFurry/steam-go"
	"github.com/GoFurry/steam-go/api/steamuserstats"
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

	resp, err := client.API.SteamUserStats.GetPlayerAchievements(
		context.Background(),
		"76561198370695025",
		550,
		&steamuserstats.GetPlayerAchievementsOptions{Language: "en"},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("achievements: %d\n", len(resp.PlayerStats.Achievements))
	for _, achievement := range resp.PlayerStats.Achievements {
		fmt.Printf("%s: %s\n", achievement.APIName, achievement.Name)
	}
}
