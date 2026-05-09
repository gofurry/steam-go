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

	fmt.Println("== SteamUserStats.GetGlobalAchievementPercentagesForApp ==")

	global, err := client.API.SteamUserStats.GetGlobalAchievementPercentagesForApp(
		realtest.BackgroundContext(),
		uint64(realtest.DefaultAppID),
	)
	if err != nil {
		realtest.Fatalf("GetGlobalAchievementPercentagesForApp failed: %v", err)
	}
	fmt.Printf("global_achievement_count=%d\n", len(global.AchievementPercentages.Achievements))
	for i, achievement := range global.AchievementPercentages.Achievements {
		if i >= 5 {
			break
		}
		fmt.Printf("[%d] name=%s percent=%s\n", i+1, achievement.Name, achievement.Percent)
	}

	fmt.Println("== SteamUserStats.GetSchemaForGame ==")

	schema, err := client.API.SteamUserStats.GetSchemaForGame(realtest.BackgroundContext(), realtest.DefaultAppID)
	if err != nil {
		realtest.Fatalf("GetSchemaForGame failed: %v", err)
	}
	fmt.Printf("game=%s version=%s achievements=%d stats=%d\n",
		schema.Game.GameName,
		schema.Game.GameVersion,
		len(schema.Game.AvailableGameStats.Achievements),
		len(schema.Game.AvailableGameStats.Stats),
	)

	fmt.Println("== SteamUserStats.GetNumberOfCurrentPlayers ==")

	currentPlayers, err := client.API.SteamUserStats.GetNumberOfCurrentPlayers(realtest.BackgroundContext(), realtest.DefaultAppID)
	if err != nil {
		realtest.Fatalf("GetNumberOfCurrentPlayers failed: %v", err)
	}
	fmt.Printf("player_count=%d result=%d\n", currentPlayers.Response.PlayerCount, currentPlayers.Response.Result)

	fmt.Println("== SteamUserStats.GetUserStatsForGame ==")

	userStats, err := client.API.SteamUserStats.GetUserStatsForGame(
		realtest.BackgroundContext(),
		realtest.DefaultSteamID,
		realtest.DefaultAppID,
	)
	if err != nil {
		realtest.Fatalf("GetUserStatsForGame failed: %v", err)
	}
	fmt.Printf("game=%s achievements=%d stats=%d\n",
		userStats.PlayerStats.GameName,
		len(userStats.PlayerStats.Achievements),
		len(userStats.PlayerStats.Stats),
	)
}
