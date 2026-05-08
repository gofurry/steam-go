package mobilenotificationservice

// GetUserNotificationCountsResponse matches IMobileNotificationService/GetUserNotificationCounts/v1.
type GetUserNotificationCountsResponse struct {
	Response UserNotificationCounts `json:"response"`
}

// UserNotificationCounts is the top-level notification count payload.
type UserNotificationCounts struct {
	Notifications     []UserNotification `json:"notifications"`
	AccountAlertCount int                `json:"account_alert_count"`
}

// UserNotification matches one user notification counter.
type UserNotification struct {
	UserNotificationType int `json:"user_notification_type"`
	Count                int `json:"count"`
}
