package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/steamnews"
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
	fmt.Println("== SteamNews.GetNewsForApp ==")

	resp, err := client.API.SteamNews.GetNewsForApp(
		realtest.BackgroundContext(),
		realtest.DefaultAppID,
		&steamnews.GetNewsForAppOptions{
			Count:     3,
			MaxLength: 200,
			Tags:      []string{"patchnotes"},
		},
	)
	if err != nil {
		realtest.Fatalf("GetNewsForApp failed: %v", err)
	}

	fmt.Printf("news_count=%d\n", resp.AppNews.Count)
	for _, item := range resp.AppNews.NewsItems {
		fmt.Printf("gid=%s title=%s\n", item.GID, item.Title)
	}
}
