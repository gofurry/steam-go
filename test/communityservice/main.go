package main

import (
	"fmt"

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
	fmt.Println("== CommunityService.GetApps ==")

	resp, err := client.API.CommunityService.GetApps(realtest.BackgroundContext(), []uint32{realtest.DefaultAppID, 570})
	if err != nil {
		realtest.Fatalf("GetApps failed: %v", err)
	}

	fmt.Printf("apps=%d\n", len(resp.Response.Apps))
	for _, app := range resp.Response.Apps {
		fmt.Printf("appid=%d name=%s\n", app.AppID, app.Name)
	}
}
