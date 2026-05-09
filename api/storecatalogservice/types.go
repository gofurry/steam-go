package storecatalogservice

// GetDevPageLinksResponse matches IStoreCatalogService/GetDevPageLinks/v1.
type GetDevPageLinksResponse struct {
	Response DevPageLinksPayload `json:"response"`
}

// DevPageLinksPayload is the top-level dev page links payload.
type DevPageLinksPayload struct {
	Links []DevPageLink `json:"links"`
}

// DevPageLink matches one developer page link row.
type DevPageLink struct {
	AppID       uint32 `json:"appid"`
	ClanSteamID string `json:"clan_steamid"`
	Relation    int    `json:"relation"`
	LinkName    string `json:"linkname"`
	JSON        string `json:"json"`
}
