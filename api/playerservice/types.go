package playerservice

// GetOwnedGamesResponse matches IPlayerService/GetOwnedGames/v1.
type GetOwnedGamesResponse struct {
	Response OwnedGames `json:"response"`
}

// OwnedGames is the top-level owned games payload.
type OwnedGames struct {
	GameCount int         `json:"game_count"`
	Games     []OwnedGame `json:"games"`
}

// OwnedGame matches the Valve owned-game schema.
type OwnedGame struct {
	AppID                  uint32 `json:"appid"`
	Name                   string `json:"name"`
	Playtime2Weeks         int    `json:"playtime_2weeks"`
	PlaytimeForever        int    `json:"playtime_forever"`
	ImgIconURL             string `json:"img_icon_url"`
	HasCommunityVisible    bool   `json:"has_community_visible_stats"`
	PlaytimeWindowsForever int    `json:"playtime_windows_forever"`
	PlaytimeMacForever     int    `json:"playtime_mac_forever"`
	PlaytimeLinuxForever   int    `json:"playtime_linux_forever"`
	PlaytimeDeckForever    int    `json:"playtime_deck_forever"`
	RTimeLastPlayed        int64  `json:"rtime_last_played"`
	CapsuleFilename        string `json:"capsule_filename"`
	HasWorkshop            bool   `json:"has_workshop"`
	HasMarket              bool   `json:"has_market"`
	HasDLC                 bool   `json:"has_dlc"`
	ContentDescriptorIDs   []int  `json:"content_descriptorids"`
	PlaytimeDisconnected   int    `json:"playtime_disconnected"`
}
