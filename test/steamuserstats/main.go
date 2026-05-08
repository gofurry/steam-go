package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/steamuserstats"
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
	fmt.Println("== SteamUserStats.GetPlayerAchievements ==")

	resp, err := client.API.SteamUserStats.GetPlayerAchievements(
		realtest.BackgroundContext(),
		realtest.DefaultSteamID,
		realtest.DefaultAppID,
		&steamuserstats.GetPlayerAchievementsOptions{Language: "zh"},
	)
	if err != nil {
		realtest.Fatalf("GetPlayerAchievements failed: %v", err)
	}

	fmt.Printf("achievement_count=%d\n", len(resp.PlayerStats.Achievements))
	for i, achievement := range resp.PlayerStats.Achievements {
		if i >= 10 {
			break
		}
		fmt.Printf("[%d] api=%s name=%s achieved=%d\n", i+1, achievement.APIName, achievement.Name, achievement.Achieved)
	}
}
