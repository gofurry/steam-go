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

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	realtest.PrintProxy(cfg)

	if !realtest.RequireAPIKey(cfg) {
		return
	}

	fmt.Println("== StoreTopSellersService.GetCountryList ==")

	countries, err := client.API.StoreTopSellersService.GetCountryList(realtest.BackgroundContext())
	if err != nil {
		realtest.Fatalf("GetCountryList failed: %v", err)
	}
	fmt.Printf("countries=%d\n", len(countries.Response.Countries))
	for i, country := range countries.Response.Countries {
		if i >= 10 {
			break
		}
		fmt.Printf("[%d] %s %s\n", i+1, country.CountryCode, country.Name)
	}

	fmt.Println("== StoreTopSellersService.GetWeeklyTopSellers ==")

	weekly, err := client.API.StoreTopSellersService.GetWeeklyTopSellers(realtest.BackgroundContext(), "CN")
	if err != nil {
		realtest.Fatalf("GetWeeklyTopSellers failed: %v", err)
	}
	fmt.Printf("weekly_payload_bytes=%d\n", len(weekly.Response))
}
