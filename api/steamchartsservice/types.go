package steamchartsservice

// GetBestOfYearPagesResponse matches ISteamChartsService/GetBestOfYearPages/v1.
type GetBestOfYearPagesResponse struct {
	Response BestOfYearPagesPayload `json:"response"`
}

// BestOfYearPagesPayload is the top-level yearly page listing payload.
type BestOfYearPagesPayload struct {
	Pages []BestOfYearPage `json:"pages"`
}

// BestOfYearPage matches one Best of Steam yearly page entry.
type BestOfYearPage struct {
	Name            string   `json:"name"`
	URLPath         string   `json:"url_path"`
	BannerURL       []string `json:"banner_url"`
	BannerURLMobile []string `json:"banner_url_mobile"`
	StartDate       int64    `json:"start_date"`
}

// GetGamesByConcurrentPlayersResponse matches ISteamChartsService/GetGamesByConcurrentPlayers/v1.
type GetGamesByConcurrentPlayersResponse struct {
	Response GamesByConcurrentPlayersPayload `json:"response"`
}

// GamesByConcurrentPlayersPayload is the top-level concurrent-player leaderboard payload.
type GamesByConcurrentPlayersPayload struct {
	LastUpdate int64                          `json:"last_update"`
	Ranks      []GamesByConcurrentPlayersRank `json:"ranks"`
}

// GamesByConcurrentPlayersRank matches one rank entry in GetGamesByConcurrentPlayers.
type GamesByConcurrentPlayersRank struct {
	Rank             int    `json:"rank"`
	AppID            uint32 `json:"appid"`
	ConcurrentInGame int    `json:"concurrent_in_game"`
	PeakInGame       int    `json:"peak_in_game"`
}

// GetMonthTopAppReleasesResponse matches ISteamChartsService/GetMonthTopAppReleases/v1.
type GetMonthTopAppReleasesResponse struct {
	Response MonthTopAppReleasesPayload `json:"response"`
}

// MonthTopAppReleasesPayload is the top-level monthly release ranking payload.
type MonthTopAppReleasesPayload struct {
	TopDLCReleases               []TopAppRelease `json:"top_dlc_releases"`
	TopCombinedAppAndDLCReleases []TopAppRelease `json:"top_combined_app_and_dlc_releases"`
}

// TopAppRelease matches one release ranking entry.
type TopAppRelease struct {
	AppID          uint32 `json:"appid"`
	RTimeRelease   int64  `json:"rtime_release"`
	AppReleaseRank int    `json:"app_release_rank"`
}

// GetMostPlayedGamesResponse matches ISteamChartsService/GetMostPlayedGames/v1.
type GetMostPlayedGamesResponse struct {
	Response MostPlayedGamesPayload `json:"response"`
}

// MostPlayedGamesPayload is the top-level most-played ranking payload.
type MostPlayedGamesPayload struct {
	RollupDate int64                `json:"rollup_date"`
	Ranks      []MostPlayedGameRank `json:"ranks"`
}

// MostPlayedGameRank matches one rank entry in GetMostPlayedGames.
type MostPlayedGameRank struct {
	Rank         int    `json:"rank"`
	AppID        uint32 `json:"appid"`
	LastWeekRank int    `json:"last_week_rank"`
	PeakInGame   int    `json:"peak_in_game"`
}

// GetTopReleasesPagesResponse matches ISteamChartsService/GetTopReleasesPages/v1.
type GetTopReleasesPagesResponse struct {
	Response TopReleasesPagesPayload `json:"response"`
}

// TopReleasesPagesPayload is the top-level monthly release page listing payload.
type TopReleasesPagesPayload struct {
	Pages []TopReleasesPage `json:"pages"`
}

// TopReleasesPage matches one top-releases landing page entry.
type TopReleasesPage struct {
	Name         string        `json:"name"`
	StartOfMonth int64         `json:"start_of_month"`
	URLPath      string        `json:"url_path"`
	ItemIDs      []ChartAppRef `json:"item_ids"`
}

// ChartAppRef wraps app identifiers embedded in chart payloads.
type ChartAppRef struct {
	AppID uint32 `json:"appid"`
}

// GetYearTopAppReleasesResponse matches ISteamChartsService/GetYearTopAppReleases/v1.
type GetYearTopAppReleasesResponse struct {
	Response YearTopAppReleasesPayload `json:"response"`
}

// YearTopAppReleasesPayload is the top-level yearly release ranking payload.
type YearTopAppReleasesPayload struct {
	TopDLCReleases               []TopAppRelease      `json:"top_dlc_releases"`
	TopCombinedAppAndDLCReleases []TopAppRelease      `json:"top_combined_app_and_dlc_releases"`
	TopAppList                   []YearTopAppListItem `json:"top_app_list"`
}

// YearTopAppListItem matches one entry in the yearly top-app rollup list.
type YearTopAppListItem struct {
	AppID          uint32 `json:"appid"`
	AppReleaseRank int    `json:"app_release_rank"`
	Type           int    `json:"type"`
}
