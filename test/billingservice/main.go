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
	if !realtest.RequireAccessToken(cfg) {
		return
	}

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	realtest.PrintProxy(cfg)
	fmt.Println("== BillingService.GetRecurringSubscriptionsCount ==")

	resp, err := client.API.BillingService.GetRecurringSubscriptionsCount(realtest.BackgroundContext())
	if err != nil {
		realtest.Fatalf("GetRecurringSubscriptionsCount failed: %v", err)
	}

	fmt.Printf("active=%d inactive=%d\n", resp.Response.ActiveSubscriptionsCount, resp.Response.InactiveSubscriptionsCount)
}
