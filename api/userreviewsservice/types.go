package userreviewsservice

// GetFriendsRecommendedAppResponse matches IUserReviewsService/GetFriendsRecommendedApp/v1.
type GetFriendsRecommendedAppResponse struct {
	Response FriendsRecommendedAppPayload `json:"response"`
}

// FriendsRecommendedAppPayload is the top-level friends recommendation payload.
type FriendsRecommendedAppPayload struct {
	AccountIDsRecommended []uint32 `json:"accountids_recommended"`
}
