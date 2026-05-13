package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/steamnotificationservice"
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
	fmt.Println("== SteamNotificationService.GetPreferences ==")

	preferences, err := client.API.SteamNotificationService.GetPreferences(realtest.BackgroundContext(), cfg.AccessToken)
	if err != nil {
		realtest.Fatalf("GetPreferences failed: %v", err)
	}
	fmt.Printf("preferences=%d\n", len(preferences.Response.Preferences))
	for _, preference := range preferences.Response.Preferences {
		fmt.Printf("type=%d targets=%d\n", preference.NotificationType, preference.NotificationTargets)
	}

	fmt.Println("== SteamNotificationService.GetSteamNotifications ==")

	trueValue := true
	notifications, err := client.API.SteamNotificationService.GetSteamNotifications(
		realtest.BackgroundContext(),
		cfg.AccessToken,
		&steamnotificationservice.GetSteamNotificationsOptions{
			IncludeHidden:            &trueValue,
			IncludeConfirmationCount: &trueValue,
			IncludePinnedCounts:      &trueValue,
			IncludeRead:              &trueValue,
		},
	)
	if err != nil {
		realtest.Fatalf("GetSteamNotifications failed: %v", err)
	}

	fmt.Printf(
		"notifications=%d unread=%d confirmations=%d pending_gifts=%d pending_friends=%d\n",
		len(notifications.Response.Notifications),
		notifications.Response.UnreadCount,
		notifications.Response.ConfirmationCount,
		notifications.Response.PendingGiftCount,
		notifications.Response.PendingFriendCount,
	)
	for _, notification := range notifications.Response.Notifications {
		fmt.Printf("id=%s type=%d read=%t hidden=%t\n", notification.NotificationID, notification.NotificationType, notification.Read, notification.Hidden)
	}
}
