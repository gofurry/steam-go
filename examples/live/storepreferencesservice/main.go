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
	if !realtest.RequireAccessToken(cfg) {
		return
	}

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	realtest.PrintProxy(cfg)
	fmt.Println("== StorePreferencesService.GetIgnoreList ==")

	resp, err := client.API.StorePreferencesService.GetIgnoreList(realtest.BackgroundContext(), cfg.AccessToken)
	if err != nil {
		realtest.Fatalf("GetIgnoreList failed: %v", err)
	}

	fmt.Printf("ignored_apps=%d\n", len(resp.Response.IgnoreList))
	for i, app := range resp.Response.IgnoreList {
		if i >= 20 {
			break
		}
		fmt.Printf("[%d] appid=%d reason=%d\n", i+1, app.AppID, app.Reason)
	}
}
