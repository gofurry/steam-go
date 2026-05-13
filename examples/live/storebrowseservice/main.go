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
	fmt.Println("== StoreBrowseService.GetContentHubConfig ==")

	resp, err := client.API.StoreBrowseService.GetContentHubConfig(realtest.BackgroundContext())
	if err != nil {
		realtest.Fatalf("GetContentHubConfig failed: %v", err)
	}

	fmt.Printf("hubconfigs=%d\n", len(resp.Response.HubConfigs))
	for i, hub := range resp.Response.HubConfigs {
		if i >= 10 {
			break
		}
		fmt.Printf("[%d] id=%d handle=%s type=%s\n", i+1, hub.HubCategoryID, hub.Handle, hub.Type)
	}
}
