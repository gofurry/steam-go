package storeservice

// GetAppListResponse matches IStoreService/GetAppList/v1.
type GetAppListResponse struct {
	Response AppListPayload `json:"response"`
}

// AppListPayload is the top-level app list payload.
type AppListPayload struct {
	Apps            []StoreApp `json:"apps"`
	HaveMoreResults bool       `json:"have_more_results"`
	LastAppID       uint32     `json:"last_appid"`
}

// StoreApp matches one app list row.
type StoreApp struct {
	AppID             uint32 `json:"appid"`
	Name              string `json:"name"`
	LastModified      int64  `json:"last_modified"`
	PriceChangeNumber uint64 `json:"price_change_number"`
}

// GetGamesFollowedResponse matches IStoreService/GetGamesFollowed/v1.
type GetGamesFollowedResponse struct {
	Response GamesFollowedPayload `json:"response"`
}

// GamesFollowedPayload is the top-level followed games payload.
type GamesFollowedPayload struct {
	AppIDs []uint32 `json:"appids"`
}

// GetGamesFollowedCountResponse matches IStoreService/GetGamesFollowedCount/v1.
type GetGamesFollowedCountResponse struct {
	Response GamesFollowedCountPayload `json:"response"`
}

// GamesFollowedCountPayload is the top-level followed game count payload.
type GamesFollowedCountPayload struct {
	FollowedGameCount int `json:"followed_game_count"`
}

// GetMostPopularTagsResponse matches IStoreService/GetMostPopularTags/v1.
type GetMostPopularTagsResponse struct {
	Response MostPopularTagsPayload `json:"response"`
}

// MostPopularTagsPayload is the top-level tag list payload.
type MostPopularTagsPayload struct {
	Tags []StoreTag `json:"tags"`
}

// StoreTag matches one store tag row.
type StoreTag struct {
	TagID uint32 `json:"tagid"`
	Name  string `json:"name"`
}

// GetUserGameInterestStateResponse matches IStoreService/GetUserGameInterestState/v1.
type GetUserGameInterestStateResponse struct {
	Response UserGameInterestStatePayload `json:"response"`
}

// UserGameInterestStatePayload is the top-level interest state payload.
type UserGameInterestStatePayload struct {
	Owned               bool                `json:"owned"`
	Following           bool                `json:"following"`
	InQueues            []int               `json:"in_queues"`
	QueueItemsRemaining []int               `json:"queue_items_remaining"`
	QueueItemsNextAppID []uint32            `json:"queue_items_next_appid"`
	Queues              []UserInterestQueue `json:"queues"`
}

// UserInterestQueue matches one queue status row.
type UserInterestQueue struct {
	Type               int    `json:"type"`
	Skipped            bool   `json:"skipped"`
	ItemsRemaining     int    `json:"items_remaining"`
	NextAppID          uint32 `json:"next_appid"`
	ExperimentalCohort int    `json:"experimental_cohort"`
}
