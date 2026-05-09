package steamuseroauth

import "github.com/GoFurry/steam-go/api/steamuser"

// GetUserSummariesResponse matches ISteamUserOAuth/GetUserSummaries/v1.
type GetUserSummariesResponse struct {
	Players []steamuser.Player `json:"players"`
}

// GetFriendListResponse matches ISteamUserOAuth/GetFriendList/v1.
type GetFriendListResponse struct {
	Friends []steamuser.Friend `json:"friends"`
}
