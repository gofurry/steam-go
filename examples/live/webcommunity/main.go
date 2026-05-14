package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofurry/steam-go/examples/live/internal/realtest"
)

func main() {
	publicInventoryID := strings.TrimSpace(os.Getenv("STEAM_PUBLIC_INVENTORY_ID"))
	if publicInventoryID == "" {
		fmt.Println("skip: STEAM_PUBLIC_INVENTORY_ID is empty")
		return
	}

	cfg, err := realtest.LoadConfig()
	if err != nil {
		realtest.Fatalf("load config: %v", err)
	}
	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("new client: %v", err)
	}
	defer client.Close()

	ctx := realtest.BackgroundContext()
	realtest.PrintProxy(cfg)

	resp, err := client.Web.Community.GetInventory(ctx, publicInventoryID, 753, "6", nil)
	if err != nil {
		realtest.Fatalf("GetInventory: %v", err)
	}
	fmt.Printf("assets=%d descriptions=%d total=%d more=%v\n", len(resp.Assets), len(resp.Descriptions), resp.TotalInventoryCount, resp.MoreItems.Bool())
}
