package main

import (
	"fmt"

	"github.com/gofurry/steam-go/examples/live/internal/realtest"
)

func main() {
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

	price, err := client.Web.Market.GetPriceOverview(ctx, 440, "Mann Co. Supply Crate Key", nil)
	if err != nil {
		realtest.Fatalf("GetPriceOverview: %v", err)
	}
	fmt.Printf("lowest=%s median=%s volume=%s\n", price.LowestPrice, price.MedianPrice, price.Volume)
}
