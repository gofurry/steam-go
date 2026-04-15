package main

import (
	"context"
	"fmt"
	"os"
	"time"

	steam "github.com/GoFurry/steam-go"
	"github.com/GoFurry/steam-go/api/steamnews"
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

	resp, err := client.SteamNews.GetNewsForApp(
		context.Background(),
		570,
		&steamnews.GetNewsForAppOptions{Count: 3},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("news items: %d\n", resp.AppNews.Count)
	for _, item := range resp.AppNews.NewsItems {
		fmt.Printf("%s: %s\n", item.GID, item.Title)
	}
}
