package storepreferencesservice

// GetIgnoreListResponse matches IStorePreferencesService/GetIgnoreList/v1.
type GetIgnoreListResponse struct {
	Response IgnoreListPayload `json:"response"`
}

// IgnoreListPayload is the top-level ignore list payload.
type IgnoreListPayload struct {
	IgnoreList []IgnoredApp `json:"ignore_list"`
}

// IgnoredApp matches one ignored app row.
type IgnoredApp struct {
	AppID  uint32 `json:"appid"`
	Reason int    `json:"reason"`
}
