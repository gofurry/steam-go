package market

// PriceOverviewResponse is the typed Steam Market price overview payload.
type PriceOverviewResponse struct {
	Success     bool   `json:"success"`
	LowestPrice string `json:"lowest_price"`
	Volume      string `json:"volume"`
	MedianPrice string `json:"median_price"`
}
