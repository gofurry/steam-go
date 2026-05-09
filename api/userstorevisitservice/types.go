package userstorevisitservice

import "encoding/json"

// GetFrequentlyVisitedPagesResponse matches IUserStoreVisitService/GetFrequentlyVisitedPages/v1.
type GetFrequentlyVisitedPagesResponse struct {
	Response FrequentlyVisitedPagesPayload `json:"response"`
}

// FrequentlyVisitedPagesPayload is the top-level frequent visit payload.
type FrequentlyVisitedPagesPayload struct {
	VisitData    VisitData          `json:"visit_data"`
	FrequentHubs []FrequentHubVisit `json:"frequent_hubs"`
}

// VisitData is the visit-data section of the frequent visit payload.
type VisitData struct {
	RecentApps []RecentStoreVisit `json:"recent_apps"`
}

// RecentStoreVisit matches one recent app visit row.
type RecentStoreVisit struct {
	ItemID    StoreItemID `json:"item_id"`
	TimeVisit int64       `json:"time_visit"`
}

// FrequentHubVisit matches one frequent hub visit row.
type FrequentHubVisit struct {
	ItemID     StoreItemID `json:"item_id"`
	TimeVisit  int64       `json:"time_visit"`
	VisitCount int         `json:"visit_count"`
}

// StoreItemID matches the flexible item_id shape returned by store visit endpoints.
type StoreItemID struct {
	AppID         uint32 `json:"appid,omitempty"`
	TagID         uint32 `json:"tagid,omitempty"`
	HubCategoryID uint32 `json:"hubcategoryid,omitempty"`
}

// GetMostVisitedItemsOnStoreResponse matches IUserStoreVisitService/GetMostVisitedItemsOnStore/v1.
type GetMostVisitedItemsOnStoreResponse struct {
	Response MostVisitedItemsOnStorePayload `json:"response"`
}

// MostVisitedItemsOnStorePayload is the top-level most-visited items payload.
type MostVisitedItemsOnStorePayload struct {
	ItemIDs []StoreItemID     `json:"item_ids"`
	Items   []json.RawMessage `json:"items"`
}
