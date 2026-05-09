package questservice

// GetCommunityInventoryResponse matches IQuestService/GetCommunityInventory/v1.
type GetCommunityInventoryResponse struct {
	Response CommunityInventoryPayload `json:"response"`
}

// CommunityInventoryPayload is the top-level community inventory payload.
type CommunityInventoryPayload struct {
	Items []CommunityInventoryItem `json:"items"`
}

// CommunityInventoryItem matches one community inventory item entry.
type CommunityInventoryItem struct {
	CommunityItemID string                        `json:"communityitemid"`
	ItemType        int                           `json:"item_type"`
	AppID           uint32                        `json:"appid"`
	Owner           uint32                        `json:"owner"`
	Attributes      []CommunityInventoryAttribute `json:"attributes"`
	Used            bool                          `json:"used"`
	OwnerOrigin     int                           `json:"owner_origin"`
	Amount          string                        `json:"amount"`
}

// CommunityInventoryAttribute matches one inventory item attribute.
type CommunityInventoryAttribute struct {
	AttributeID int    `json:"attributeid"`
	Value       string `json:"value"`
}

// GetNumTradingCardsEarnedResponse matches IQuestService/GetNumTradingCardsEarned/v1.
type GetNumTradingCardsEarnedResponse struct {
	Response NumTradingCardsEarnedPayload `json:"response"`
}

// NumTradingCardsEarnedPayload is the top-level trading-card count payload.
type NumTradingCardsEarnedPayload struct {
	NumTradingCards int `json:"num_trading_cards"`
}
