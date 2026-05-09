package steamuser

// GetFriendListResponse matches ISteamUser/GetFriendList/v1.
type GetFriendListResponse struct {
	FriendsList FriendsList `json:"friendslist"`
}

// FriendsList is the top-level friend list payload.
type FriendsList struct {
	Friends []Friend `json:"friends"`
}

// Friend matches one friend relationship row.
type Friend struct {
	SteamID      string `json:"steamid"`
	Relationship string `json:"relationship"`
	FriendSince  int64  `json:"friend_since"`
}

// GetPlayerBansResponse matches ISteamUser/GetPlayerBans/v1.
type GetPlayerBansResponse struct {
	Players []PlayerBan `json:"players"`
}

// PlayerBan matches Valve's player ban structure.
type PlayerBan struct {
	SteamID          string `json:"SteamId"`
	CommunityBanned  bool   `json:"CommunityBanned"`
	VACBanned        bool   `json:"VACBanned"`
	NumberOfVACBans  int    `json:"NumberOfVACBans"`
	DaysSinceLastBan int    `json:"DaysSinceLastBan"`
	NumberOfGameBans int    `json:"NumberOfGameBans"`
	EconomyBan       string `json:"EconomyBan"`
}

// GetPlayerSummariesResponse matches ISteamUser/GetPlayerSummaries/v2.
type GetPlayerSummariesResponse struct {
	Response struct {
		Players []Player `json:"players"`
	} `json:"response"`
}

// Player matches the Steam player payload returned by Valve.
type Player struct {
	SteamID                  string `json:"steamid"`
	CommunityVisibilityState int    `json:"communityvisibilitystate"`
	ProfileState             int    `json:"profilestate"`
	PersonaName              string `json:"personaname"`
	CommentPermission        int    `json:"commentpermission"`
	ProfileURL               string `json:"profileurl"`
	Avatar                   string `json:"avatar"`
	AvatarMedium             string `json:"avatarmedium"`
	AvatarFull               string `json:"avatarfull"`
	AvatarHash               string `json:"avatarhash"`
	LastLogoff               int64  `json:"lastlogoff"`
	PersonaState             int    `json:"personastate"`
	PrimaryClanID            string `json:"primaryclanid"`
	TimeCreated              int64  `json:"timecreated"`
	PersonaStateFlags        int    `json:"personastateflags"`
	RealName                 string `json:"realname"`
	LocCountryCode           string `json:"loccountrycode"`
	LocStateCode             string `json:"locstatecode"`
	LocCityID                int    `json:"loccityid"`
}

// GetUserGroupListResponse matches ISteamUser/GetUserGroupList/v1.
type GetUserGroupListResponse struct {
	Response UserGroupList `json:"response"`
}

// UserGroupList is the top-level group list payload.
type UserGroupList struct {
	Success bool         `json:"success"`
	Groups  []SteamGroup `json:"groups"`
}

// SteamGroup matches one group ID row.
type SteamGroup struct {
	GID string `json:"gid"`
}
