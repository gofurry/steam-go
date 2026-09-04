package storebrowseservice

import (
	"bytes"
	"encoding/json"
)

// GetContentHubConfigResponse matches IStoreBrowseService/GetContentHubConfig/v1.
type GetContentHubConfigResponse struct {
	Response ContentHubConfigPayload `json:"response"`
}

// ContentHubConfigPayload is the top-level content hub config payload.
type ContentHubConfigPayload struct {
	HubConfigs []HubConfig `json:"hubconfigs"`
}

// HubConfig matches one content hub config row.
type HubConfig struct {
	HubCategoryID   uint32   `json:"hubcategoryid"`
	Type            string   `json:"type"`
	Handle          string   `json:"handle"`
	DisplayName     string   `json:"display_name"`
	URLPath         string   `json:"url_path"`
	ReplacesTags    []uint32 `json:"replaces_tags"`
	MustHaveTags    []uint32 `json:"must_have_tags"`
	AnyOneOfTags    []uint32 `json:"any_one_of_tags"`
	MustNotHaveTags []uint32 `json:"must_not_have_tags"`
	HubDescription  string   `json:"hub_description"`
}

// GetItemsRequest controls IStoreBrowseService/GetItems/v1 input_json.
type GetItemsRequest struct {
	IDs         []StoreItemID           `json:"ids,omitempty"`
	Context     *StoreBrowseContext     `json:"context,omitempty"`
	DataRequest *StoreBrowseDataRequest `json:"data_request,omitempty"`
}

// StoreItemID identifies one Store item lookup target.
type StoreItemID struct {
	AppID     uint32 `json:"appid,omitempty"`
	PackageID uint32 `json:"packageid,omitempty"`
	BundleID  uint32 `json:"bundleid,omitempty"`
}

// StoreBrowseContext controls region and language for StoreBrowse responses.
type StoreBrowseContext struct {
	CountryCode string `json:"country_code,omitempty"`
	Language    string `json:"language,omitempty"`
}

// StoreBrowseDataRequest controls optional Store item payload sections.
type StoreBrowseDataRequest struct {
	IncludeAssets bool `json:"include_assets,omitempty"`
}

// GetItemsResponse matches IStoreBrowseService/GetItems/v1.
type GetItemsResponse struct {
	Response GetItemsPayload `json:"response"`
}

// GetItemsPayload is the top-level GetItems response payload.
type GetItemsPayload struct {
	StoreItems []StoreItem `json:"store_items"`
}

// StoreItem matches one StoreBrowse item row. Large or unstable subtrees stay
// raw so the typed surface can remain compatible as Steam changes the payload.
type StoreItem struct {
	ItemType     int             `json:"item_type,omitempty"`
	ID           uint32          `json:"id,omitempty"`
	Success      int             `json:"success,omitempty"`
	Visible      bool            `json:"visible,omitempty"`
	Name         string          `json:"name,omitempty"`
	StoreURLPath string          `json:"store_url_path,omitempty"`
	AppID        uint32          `json:"appid,omitempty"`
	Type         int             `json:"type,omitempty"`
	Assets       StoreItemAssets `json:"assets,omitempty"`

	RelatedItems json.RawMessage `json:"related_items,omitempty"`
	Categories   json.RawMessage `json:"categories,omitempty"`
}

// StoreItemAssets contains Steam's asset metadata for one Store item.
//
// Steam may add keys over time, so assets are intentionally exposed as a map.
// Only string values used for asset URL discovery are retained; non-string
// metadata is ignored.
type StoreItemAssets map[string]string

// UnmarshalJSON decodes Store item assets as a string-field view of Steam's
// asset metadata. Unknown non-string metadata is ignored so upstream additions
// do not prevent known asset filenames from being discovered.
func (a *StoreItemAssets) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if fields == nil {
		*a = nil
		return nil
	}

	assets := *a
	if assets == nil {
		assets = make(StoreItemAssets, len(fields))
	}
	for key, rawValue := range fields {
		rawValue = bytes.TrimSpace(rawValue)
		if len(rawValue) == 0 || rawValue[0] != '"' {
			delete(assets, key)
			continue
		}

		var value string
		if err := json.Unmarshal(rawValue, &value); err != nil {
			return err
		}
		assets[key] = value
	}
	*a = assets
	return nil
}
