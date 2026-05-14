package community

import (
	"bytes"
	"encoding/json"
	"strings"
)

// InventoryResponse is the typed Steam Community inventory payload.
type InventoryResponse struct {
	Success             int              `json:"success"`
	Assets              []InventoryAsset `json:"assets"`
	Descriptions        []InventoryItem  `json:"descriptions"`
	TotalInventoryCount int              `json:"total_inventory_count"`
	MoreItems           FlexibleBool     `json:"more_items"`
	LastAssetID         string           `json:"last_assetid"`
	Error               string           `json:"error"`
}

// InventoryAsset is one owned asset row.
type InventoryAsset struct {
	AppID      uint32 `json:"appid"`
	ContextID  string `json:"contextid"`
	AssetID    string `json:"assetid"`
	ClassID    string `json:"classid"`
	InstanceID string `json:"instanceid"`
	Amount     string `json:"amount"`
}

// InventoryItem is the typed stable subset of one description row.
type InventoryItem struct {
	AppID             uint32            `json:"appid"`
	ClassID           string            `json:"classid"`
	InstanceID        string            `json:"instanceid"`
	Currency          int               `json:"currency"`
	Name              string            `json:"name"`
	MarketName        string            `json:"market_name"`
	MarketHashName    string            `json:"market_hash_name"`
	Type              string            `json:"type"`
	IconURL           string            `json:"icon_url"`
	IconURLLarge      string            `json:"icon_url_large"`
	NameColor         string            `json:"name_color"`
	BackgroundColor   string            `json:"background_color"`
	Tradable          int               `json:"tradable"`
	Marketable        int               `json:"marketable"`
	Commodity         int               `json:"commodity"`
	Tags              []InventoryTag    `json:"tags,omitempty"`
	Descriptions      []json.RawMessage `json:"descriptions,omitempty"`
	OwnerDescriptions []json.RawMessage `json:"owner_descriptions,omitempty"`
	Actions           []json.RawMessage `json:"actions,omitempty"`
	OwnerActions      []json.RawMessage `json:"owner_actions,omitempty"`
	MarketActions     []json.RawMessage `json:"market_actions,omitempty"`
	FraudWarnings     []json.RawMessage `json:"fraudwarnings,omitempty"`
}

// InventoryTag is one typed tag row.
type InventoryTag struct {
	Category              string `json:"category"`
	InternalName          string `json:"internal_name"`
	LocalizedCategoryName string `json:"localized_category_name"`
	LocalizedTagName      string `json:"localized_tag_name"`
	Color                 string `json:"color"`
}

// FlexibleBool accepts bool and 0/1 integer JSON payloads.
type FlexibleBool bool

// Bool reports the underlying bool value.
func (b FlexibleBool) Bool() bool {
	return bool(b)
}

// UnmarshalJSON decodes bool or 0/1 style flags.
func (b *FlexibleBool) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	switch trimmed {
	case "true", "1":
		*b = true
		return nil
	case "false", "0", "null", "":
		*b = false
		return nil
	}

	var boolValue bool
	if err := json.Unmarshal(data, &boolValue); err == nil {
		*b = FlexibleBool(boolValue)
		return nil
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var number json.Number
	if err := decoder.Decode(&number); err == nil {
		if number.String() == "1" {
			*b = true
			return nil
		}
		if number.String() == "0" {
			*b = false
			return nil
		}
	}

	return &json.UnmarshalTypeError{Value: trimmed, Type: nil}
}
