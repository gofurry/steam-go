package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/accountcartservice"
	"github.com/GoFurry/steam-go/test/internal/realtest"
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
	fmt.Println("== AccountCartService.GetCart ==")

	resp, err := client.API.AccountCartService.GetCart(
		realtest.BackgroundContext(),
		&accountcartservice.GetCartOptions{UserCountry: "cn"},
	)
	if err != nil {
		realtest.Fatalf("GetCart failed: %v", err)
	}

	fmt.Printf("cart_items=%d subtotal=%s\n", len(resp.Response.Cart.LineItems), resp.Response.Cart.Subtotal.FormattedAmount)
}
