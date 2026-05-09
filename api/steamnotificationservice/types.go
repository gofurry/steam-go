package steamnotificationservice

// GetPreferencesResponse matches ISteamNotificationService/GetPreferences/v1.
type GetPreferencesResponse struct {
	Response PreferencesPayload `json:"response"`
}

// PreferencesPayload is the top-level notification preferences payload.
type PreferencesPayload struct {
	Preferences []NotificationPreference `json:"preferences"`
}

// NotificationPreference describes one notification preference row.
type NotificationPreference struct {
	NotificationType    int `json:"notification_type"`
	NotificationTargets int `json:"notification_targets"`
}

// GetSteamNotificationsResponse matches ISteamNotificationService/GetSteamNotifications/v1.
type GetSteamNotificationsResponse struct {
	Response SteamNotificationsPayload `json:"response"`
}

// SteamNotificationsPayload is the top-level notifications payload.
type SteamNotificationsPayload struct {
	Notifications            []SteamNotification `json:"notifications"`
	ConfirmationCount        int                 `json:"confirmation_count"`
	PendingGiftCount         int                 `json:"pending_gift_count"`
	PendingFriendCount       int                 `json:"pending_friend_count"`
	UnreadCount              int                 `json:"unread_count"`
	PendingFamilyInviteCount int                 `json:"pending_family_invite_count"`
}

// SteamNotification matches one Steam notification entry.
type SteamNotification struct {
	NotificationID      string `json:"notification_id"`
	NotificationTargets int    `json:"notification_targets"`
	NotificationType    int    `json:"notification_type"`
	BodyData            string `json:"body_data"`
	Read                bool   `json:"read"`
	Timestamp           int64  `json:"timestamp"`
	Hidden              bool   `json:"hidden"`
	Expiry              int64  `json:"expiry"`
	Viewed              int64  `json:"viewed"`
}
