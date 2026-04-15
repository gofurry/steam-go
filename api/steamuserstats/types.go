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
