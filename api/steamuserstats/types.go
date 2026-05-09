package steamuserstats

// GetPlayerAchievementsResponse matches ISteamUserStats/GetPlayerAchievements/v1.
type GetPlayerAchievementsResponse struct {
	PlayerStats PlayerStats `json:"playerstats"`
}

// PlayerStats is the top-level player stats payload.
type PlayerStats struct {
	SteamID      string        `json:"steamID"`
	GameName     string        `json:"gameName"`
	Achievements []Achievement `json:"achievements"`
	Success      bool          `json:"success"`
}

// Achievement matches the Valve achievement schema.
type Achievement struct {
	APIName     string `json:"apiname"`
	Achieved    int    `json:"achieved"`
	UnlockTime  int64  `json:"unlocktime"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetGlobalAchievementPercentagesForAppResponse matches ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2.
type GetGlobalAchievementPercentagesForAppResponse struct {
	AchievementPercentages GlobalAchievementPercentages `json:"achievementpercentages"`
}

// GlobalAchievementPercentages is the top-level global percentage payload.
type GlobalAchievementPercentages struct {
	Achievements []GlobalAchievementPercentage `json:"achievements"`
}

// GlobalAchievementPercentage matches one achievement percentage row.
type GlobalAchievementPercentage struct {
	Name    string `json:"name"`
	Percent string `json:"percent"`
}

// GetSchemaForGameResponse matches ISteamUserStats/GetSchemaForGame/v2.
type GetSchemaForGameResponse struct {
	Game GameSchema `json:"game"`
}

// GameSchema is the top-level game schema payload.
type GameSchema struct {
	GameName           string             `json:"gameName"`
	GameVersion        string             `json:"gameVersion"`
	AvailableGameStats AvailableGameStats `json:"availableGameStats"`
}

// AvailableGameStats groups achievement and stat schema metadata.
type AvailableGameStats struct {
	Achievements []GameSchemaAchievement `json:"achievements"`
	Stats        []GameSchemaStat        `json:"stats"`
}

// GameSchemaAchievement matches one achievement schema row.
type GameSchemaAchievement struct {
	Name         string  `json:"name"`
	DefaultValue float64 `json:"defaultvalue"`
	DisplayName  string  `json:"displayName"`
	Hidden       int     `json:"hidden"`
	Description  string  `json:"description"`
	Icon         string  `json:"icon"`
	IconGray     string  `json:"icongray"`
}

// GameSchemaStat matches one stat schema row.
type GameSchemaStat struct {
	Name         string  `json:"name"`
	DefaultValue float64 `json:"defaultvalue"`
	DisplayName  string  `json:"displayName"`
}

// GetNumberOfCurrentPlayersResponse matches ISteamUserStats/GetNumberOfCurrentPlayers/v1.
type GetNumberOfCurrentPlayersResponse struct {
	Response NumberOfCurrentPlayers `json:"response"`
}

// NumberOfCurrentPlayers is the top-level player-count payload.
type NumberOfCurrentPlayers struct {
	PlayerCount int `json:"player_count"`
	Result      int `json:"result"`
}

// GetUserStatsForGameResponse matches ISteamUserStats/GetUserStatsForGame/v2.
type GetUserStatsForGameResponse struct {
	PlayerStats UserStatsForGamePlayerStats `json:"playerstats"`
}

// UserStatsForGamePlayerStats is the top-level user stats payload.
type UserStatsForGamePlayerStats struct {
	SteamID      string                        `json:"steamID"`
	GameName     string                        `json:"gameName"`
	Achievements []UserStatsForGameAchievement `json:"achievements"`
	Stats        []UserStatsForGameStat        `json:"stats"`
}

// UserStatsForGameAchievement matches one achievement flag row.
type UserStatsForGameAchievement struct {
	Name     string `json:"name"`
	Achieved int    `json:"achieved"`
}

// UserStatsForGameStat matches one stat row.
type UserStatsForGameStat struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}
