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

// GetUserYearInReviewResponse matches ISaleFeatureService/GetUserYearInReview/v1.
type GetUserYearInReviewResponse struct {
	Response UserYearInReviewPayload `json:"response"`
}

// UserYearInReviewPayload is the top-level year-in-review payload.
type UserYearInReviewPayload struct {
	Stats            UserYearInReviewStats            `json:"stats"`
	PrivacyState     int                              `json:"privacy_state"`
	PerformanceStats UserYearInReviewPerformanceStats `json:"performance_stats"`
	Distribution     UserYearInReviewDistribution     `json:"distribution"`
}

// UserYearInReviewStats holds the core Year in Review aggregates.
type UserYearInReviewStats struct {
	AccountID       uint32                        `json:"account_id"`
	Year            uint32                        `json:"year"`
	PlaytimeStats   UserYearInReviewPlaytimeStats `json:"playtime_stats"`
	DemosPlayed     int                           `json:"demos_played"`
	GameRankings    UserYearInReviewGameRankings  `json:"game_rankings"`
	PlaytestsPlayed int                           `json:"playtests_played"`
	SummaryStats    UserYearInReviewSummaryStats  `json:"summary_stats"`
	Substantial     bool                          `json:"substantial"`
	TagStats        UserYearInReviewTagStats      `json:"tag_stats"`
	ByNumbers       UserYearInReviewByNumbers     `json:"by_numbers"`
}

type UserYearInReviewPlaytimeStats struct {
	TotalStats     UserYearInReviewPlaytimeBreakdown `json:"total_stats"`
	Games          []UserYearInReviewGame            `json:"games"`
	PlaytimeStreak UserYearInReviewPlaytimeStreak    `json:"playtime_streak"`
	Months         []UserYearInReviewMonth           `json:"months"`
	GameSummary    []UserYearInReviewGameSummary     `json:"game_summary"`
}

type UserYearInReviewPlaytimeBreakdown struct {
	TotalPlaytimeSeconds             int `json:"total_playtime_seconds"`
	TotalSessions                    int `json:"total_sessions"`
	VRSessions                       int `json:"vr_sessions"`
	DeckSessions                     int `json:"deck_sessions"`
	ControllerSessions               int `json:"controller_sessions"`
	LinuxSessions                    int `json:"linux_sessions"`
	MacOSSessions                    int `json:"macos_sessions"`
	WindowsSessions                  int `json:"windows_sessions"`
	TotalPlaytimePercentageX100      int `json:"total_playtime_percentagex100"`
	VRPlaytimePercentageX100         int `json:"vr_playtime_percentagex100"`
	DeckPlaytimePercentageX100       int `json:"deck_playtime_percentagex100"`
	ControllerPlaytimePercentageX100 int `json:"controller_playtime_percentagex100"`
	LinuxPlaytimePercentageX100      int `json:"linux_playtime_percentagex100"`
	MacOSPlaytimePercentageX100      int `json:"macos_playtime_percentagex100"`
	WindowsPlaytimePercentageX100    int `json:"windows_playtime_percentagex100"`
}

type UserYearInReviewGame struct {
	AppID             uint32                            `json:"appid"`
	Stats             UserYearInReviewPlaytimeBreakdown `json:"stats"`
	PlaytimeStreak    UserYearInReviewPlaytimeStreak    `json:"playtime_streak"`
	PlaytimeRanks     UserYearInReviewPlaytimeRanks     `json:"playtime_ranks"`
	RTimeFirstPlayed  int64                             `json:"rtime_first_played"`
	RelativeGameStats UserYearInReviewPlaytimeBreakdown `json:"relative_game_stats"`
}

type UserYearInReviewPlaytimeStreak struct {
	LongestConsecutiveDays int                      `json:"longest_consecutive_days"`
	RTimeStart             int64                    `json:"rtime_start"`
	StreakGames            []UserYearInReviewAppRef `json:"streak_games"`
}

type UserYearInReviewAppRef struct {
	AppID uint32 `json:"appid"`
}

type UserYearInReviewPlaytimeRanks struct {
	OverallRank int `json:"overall_rank"`
	WindowsRank int `json:"windows_rank"`
}

type UserYearInReviewMonth struct {
	RTimeMonth           int64                              `json:"rtime_month"`
	Stats                UserYearInReviewPlaytimeBreakdown  `json:"stats"`
	AppID                []UserYearInReviewMonthGame        `json:"appid"`
	RelativeMonthlyStats UserYearInReviewPlaytimeBreakdown  `json:"relative_monthly_stats"`
	GameSummary          []UserYearInReviewMonthGameSummary `json:"game_summary"`
}

type UserYearInReviewMonthGame struct {
	AppID             uint32                            `json:"appid"`
	Stats             UserYearInReviewPlaytimeBreakdown `json:"stats"`
	RTimeFirstPlayed  int64                             `json:"rtime_first_played"`
	RelativeGameStats UserYearInReviewPlaytimeBreakdown `json:"relative_game_stats"`
}

type UserYearInReviewMonthGameSummary struct {
	AppID                          uint32 `json:"appid"`
	TotalPlaytimePercentageX100    int    `json:"total_playtime_percentagex100"`
	RelativePlaytimePercentageX100 int    `json:"relative_playtime_percentagex100"`
}

type UserYearInReviewGameSummary struct {
	AppID                       uint32 `json:"appid"`
	NewThisYear                 bool   `json:"new_this_year"`
	RTimeFirstPlayedLifetime    int64  `json:"rtime_first_played_lifetime"`
	Demo                        bool   `json:"demo"`
	Playtest                    bool   `json:"playtest"`
	PlayedVR                    bool   `json:"played_vr"`
	PlayedDeck                  bool   `json:"played_deck"`
	PlayedController            bool   `json:"played_controller"`
	PlayedLinux                 bool   `json:"played_linux"`
	PlayedMac                   bool   `json:"played_mac"`
	PlayedWindows               bool   `json:"played_windows"`
	TotalPlaytimePercentageX100 int    `json:"total_playtime_percentagex100"`
	TotalSessions               int    `json:"total_sessions"`
	RTimeReleaseDate            int64  `json:"rtime_release_date"`
}

type UserYearInReviewGameRankings struct {
	OverallRanking    UserYearInReviewRankingCategory `json:"overall_ranking"`
	VRRanking         UserYearInReviewRankingCategory `json:"vr_ranking"`
	DeckRanking       UserYearInReviewRankingCategory `json:"deck_ranking"`
	ControllerRanking UserYearInReviewRankingCategory `json:"controller_ranking"`
	LinuxRanking      UserYearInReviewRankingCategory `json:"linux_ranking"`
	MacRanking        UserYearInReviewRankingCategory `json:"mac_ranking"`
	WindowsRanking    UserYearInReviewRankingCategory `json:"windows_ranking"`
}

type UserYearInReviewRankingCategory struct {
	Category string                         `json:"category"`
	Rankings []UserYearInReviewRankingEntry `json:"rankings"`
}

type UserYearInReviewRankingEntry struct {
	AppID                          uint32 `json:"appid"`
	Rank                           int    `json:"rank"`
	RelativePlaytimePercentageX100 int    `json:"relative_playtime_percentagex100"`
}

type UserYearInReviewSummaryStats struct {
	TotalAchievements          int `json:"total_achievements"`
	TotalGamesWithAchievements int `json:"total_games_with_achievements"`
	TotalRareAchievements      int `json:"total_rare_achievements"`
}

type UserYearInReviewTagStats struct {
	Stats []UserYearInReviewTagStat `json:"stats"`
}

type UserYearInReviewTagStat struct {
	TagID                 int     `json:"tag_id"`
	TagWeight             float64 `json:"tag_weight"`
	TagWeightPreSelection float64 `json:"tag_weight_pre_selection"`
}

type UserYearInReviewByNumbers struct {
	ScreenshotsShared     int `json:"screenshots_shared"`
	GiftsSent             int `json:"gifts_sent"`
	LoyaltyReactions      int `json:"loyalty_reactions"`
	WrittenReviews        int `json:"written_reviews"`
	GuidesSubmitted       int `json:"guides_submitted"`
	WorkshopContributions int `json:"workshop_contributions"`
	BadgesEarned          int `json:"badges_earned"`
	FriendsAdded          int `json:"friends_added"`
	ForumPosts            int `json:"forum_posts"`
	WorkshopSubscriptions int `json:"workshop_subscriptions"`
	GuideSubscribers      int `json:"guide_subscribers"`
	WorkshopSubscribers   int `json:"workshop_subscribers"`
	GamesPlayedPct        int `json:"games_played_pct"`
	AchievementsPct       int `json:"achievements_pct"`
	GameStreakPct         int `json:"game_streak_pct"`
	GamesPlayedAvg        int `json:"games_played_avg"`
	AchievementsAvg       int `json:"achievements_avg"`
	GameStreakAvg         int `json:"game_streak_avg"`
}

type UserYearInReviewPerformanceStats struct {
	FromDBO             bool   `json:"from_dbo"`
	OverallTimeMS       string `json:"overall_time_ms"`
	DBOLoadMS           string `json:"dbo_load_ms"`
	MessagePopulationMS string `json:"message_population_ms"`
	DBOLockLoadMS       string `json:"dbo_lock_load_ms"`
}

type UserYearInReviewDistribution struct {
	NewReleases      int `json:"new_releases"`
	RecentReleases   int `json:"recent_releases"`
	ClassicReleases  int `json:"classic_releases"`
	RecentCutoffYear int `json:"recent_cutoff_year"`
}
