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
	fmt.Println("== StoreCatalogService.GetDevPageLinks ==")

	resp, err := client.API.StoreCatalogService.GetDevPageLinks(realtest.BackgroundContext(), realtest.DefaultAppID)
	if err != nil {
		realtest.Fatalf("GetDevPageLinks failed: %v", err)
	}

	fmt.Printf("links=%d\n", len(resp.Response.Links))
	for i, link := range resp.Response.Links {
		if i >= 10 {
			break
		}
		fmt.Printf("[%d] appid=%d linkname=%s clan=%s\n", i+1, link.AppID, link.LinkName, link.ClanSteamID)
	}
}
