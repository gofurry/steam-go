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

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	realtest.PrintProxy(cfg)

	if !realtest.RequireAccessToken(cfg) {
		return
	}

	fmt.Println("== UserAccountService.GetUserCountry ==")

	resp, err := client.API.UserAccountService.GetUserCountry(
		realtest.BackgroundContext(),
		cfg.AccessToken,
		realtest.DefaultSteamID,
	)
	if err != nil {
		realtest.Fatalf("GetUserCountry failed: %v", err)
	}
	fmt.Printf("country=%s\n", resp.Response.Country)
}
