package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/questservice"
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

	ctx := realtest.BackgroundContext()
	realtest.PrintProxy(cfg)

	fmt.Println("== QuestService.GetCommunityInventory ==")
	inventoryResp, err := client.API.QuestService.GetCommunityInventory(
		ctx,
		&questservice.GetCommunityInventoryOptions{FilterAppIDs: []uint32{realtest.DefaultAppID}},
	)
	if err != nil {
		realtest.Fatalf("GetCommunityInventory failed: %v", err)
	}
	fmt.Printf("items=%d\n", len(inventoryResp.Response.Items))

	if !realtest.RequireAccessToken(cfg) {
		return
	}

	fmt.Println("\n== QuestService.GetNumTradingCardsEarned ==")
	numTradingCardsResp, err := client.API.QuestService.GetNumTradingCardsEarned(
		ctx,
		cfg.AccessToken,
		&questservice.GetNumTradingCardsEarnedOptions{},
	)
	if err != nil {
		realtest.Fatalf("GetNumTradingCardsEarned failed: %v", err)
	}
	fmt.Printf("num_trading_cards=%d\n", numTradingCardsResp.Response.NumTradingCards)
}
