package storebrowseservice

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
