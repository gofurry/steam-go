package storetopsellersservice

import "encoding/json"

// GetCountryListResponse matches IStoreTopSellersService/GetCountryList/v1.
type GetCountryListResponse struct {
	Response CountryListPayload `json:"response"`
}

// CountryListPayload is the top-level country list payload.
type CountryListPayload struct {
	Countries []Country `json:"countries"`
}

// Country matches one supported country row.
type Country struct {
	CountryCode string `json:"country_code"`
	Name        string `json:"name"`
}

// GetWeeklyTopSellersResponse matches IStoreTopSellersService/GetWeeklyTopSellers/v1.
type GetWeeklyTopSellersResponse struct {
	Response json.RawMessage `json:"response"`
}
