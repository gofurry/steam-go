package salefeatureservice

// GetFriendsSharedYearInReviewResponse matches ISaleFeatureService/GetFriendsSharedYearInReview/v1.
type GetFriendsSharedYearInReviewResponse struct {
	Response FriendsSharedYearInReviewPayload `json:"response"`
}

// FriendsSharedYearInReviewPayload is the top-level friend-share payload.
type FriendsSharedYearInReviewPayload struct {
	FriendShares []FriendYearInReviewShare `json:"friend_shares"`
}

// FriendYearInReviewShare matches one friend share entry.
type FriendYearInReviewShare struct {
	SteamID          string `json:"steamid"`
	PrivacyState     int    `json:"privacy_state"`
	RTPrivacyUpdated int64  `json:"rt_privacy_updated"`
	PrivacyOverride  bool   `json:"privacy_override"`
}

// GetUserYearAchievementsResponse matches ISaleFeatureService/GetUserYearAchievements/v1.
type GetUserYearAchievementsResponse struct {
	Response UserYearAchievementsPayload `json:"response"`
}

// UserYearAchievementsPayload is the top-level year-achievements payload.
type UserYearAchievementsPayload struct {
	GameAchievements           []UserYearAchievementGame `json:"game_achievements"`
	TotalAchievements          int                       `json:"total_achievements"`
	TotalRareAchievements      int                       `json:"total_rare_achievements"`
	TotalGamesWithAchievements int                       `json:"total_games_with_achievements"`
}

// UserYearAchievementGame matches one game entry in GetUserYearAchievements.
type UserYearAchievementGame struct {
	AppID                       uint32                `json:"appid"`
	Achievements                []UserYearAchievement `json:"achievements"`
	AllTimeUnlockedAchievements int                   `json:"all_time_unlocked_achievements"`
	UnlockedMoreInFuture        bool                  `json:"unlocked_more_in_future"`
}

// UserYearAchievement matches one achievement entry in GetUserYearAchievements.
type UserYearAchievement struct {
	StatID                  int    `json:"statid"`
	FieldID                 int    `json:"fieldid"`
	AchievementNameInternal string `json:"achievement_name_internal"`
}
