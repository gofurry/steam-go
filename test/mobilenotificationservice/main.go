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
	fmt.Println("== MobileNotificationService.GetUserNotificationCounts ==")

	resp, err := client.API.MobileNotificationService.GetUserNotificationCounts(realtest.BackgroundContext(), cfg.AccessToken)
	if err != nil {
		realtest.Fatalf("GetUserNotificationCounts failed: %v", err)
	}

	fmt.Printf("notifications=%d account_alert_count=%d\n", len(resp.Response.Notifications), resp.Response.AccountAlertCount)
	for _, notification := range resp.Response.Notifications {
		fmt.Printf("type=%d count=%d\n", notification.UserNotificationType, notification.Count)
	}
}
