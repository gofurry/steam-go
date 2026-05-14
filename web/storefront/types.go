package storefront

import (
	"encoding/json"
	"strconv"
	"strings"
)

// AppDetailsEnvelope is the keyed Storefront app details payload.
type AppDetailsEnvelope map[string]AppDetailsResult

// AppDetailsResult is one app details lookup result.
type AppDetailsResult struct {
	Success bool           `json:"success"`
	Data    AppDetailsData `json:"data"`
}

// AppDetailsData is the stable typed subset of Storefront app details.
type AppDetailsData struct {
	Type             string            `json:"type"`
	Name             string            `json:"name"`
	SteamAppID       uint32            `json:"steam_appid"`
	IsFree           bool              `json:"is_free"`
	ShortDescription string            `json:"short_description"`
	HeaderImage      string            `json:"header_image"`
	Developers       []string          `json:"developers"`
	Publishers       []string          `json:"publishers"`
	PriceOverview    *StorePrice       `json:"price_overview,omitempty"`
	Platforms        StorePlatforms    `json:"platforms"`
	Categories       []StoreCategory   `json:"categories,omitempty"`
	Genres           []StoreGenre      `json:"genres,omitempty"`
	Packages         []uint32          `json:"packages,omitempty"`
	PackageGroups    json.RawMessage   `json:"package_groups,omitempty"`
	ReleaseDate      *StoreReleaseDate `json:"release_date,omitempty"`
}

// PackageDetailsEnvelope is the keyed Storefront package details payload.
type PackageDetailsEnvelope map[string]PackageDetailsResult

// PackageDetailsResult is one package details lookup result.
type PackageDetailsResult struct {
	Success bool               `json:"success"`
	Data    PackageDetailsData `json:"data"`
}

// PackageDetailsData is the stable typed subset of Storefront package details.
type PackageDetailsData struct {
	PackageID   uint32             `json:"packageid"`
	Name        string             `json:"name"`
	HeaderImage string             `json:"header_image"`
	SmallLogo   string             `json:"small_logo"`
	PageContent string             `json:"page_content"`
	Apps        []StorePackageApp  `json:"apps,omitempty"`
	Price       *StorePackagePrice `json:"price,omitempty"`
	Platforms   StorePlatforms     `json:"platforms"`
	Categories  []StoreCategory    `json:"categories,omitempty"`
	Genres      []StoreGenre       `json:"genres,omitempty"`
	ReleaseDate *StoreReleaseDate  `json:"release_date,omitempty"`
	Details     json.RawMessage    `json:"details,omitempty"`
}

// AppReviewsResponse is the Storefront reviews payload.
type AppReviewsResponse struct {
	Success      int                     `json:"success"`
	QuerySummary StoreReviewQuerySummary `json:"query_summary"`
	Cursor       string                  `json:"cursor"`
	Reviews      []AppReview             `json:"reviews"`
}

// StoreReviewQuerySummary is the typed review summary payload.
type StoreReviewQuerySummary struct {
	NumReviews      int    `json:"num_reviews"`
	ReviewScore     int    `json:"review_score"`
	ReviewScoreDesc string `json:"review_score_desc"`
	TotalPositive   int    `json:"total_positive"`
	TotalNegative   int    `json:"total_negative"`
	TotalReviews    int    `json:"total_reviews"`
}

// AppReview is one Storefront app review.
type AppReview struct {
	RecommendationID         string          `json:"recommendationid"`
	Author                   AppReviewAuthor `json:"author"`
	Review                   string          `json:"review"`
	TimestampCreated         int64           `json:"timestamp_created"`
	TimestampUpdated         int64           `json:"timestamp_updated"`
	VotedUp                  bool            `json:"voted_up"`
	VotesUp                  int             `json:"votes_up"`
	VotesFunny               int             `json:"votes_funny"`
	WeightedVoteScore        FlexibleFloat64 `json:"weighted_vote_score"`
	SteamPurchase            bool            `json:"steam_purchase"`
	ReceivedForFree          bool            `json:"received_for_free"`
	WrittenDuringEarlyAccess bool            `json:"written_during_early_access"`
	DeveloperResponse        string          `json:"developer_response"`
	PrimarilySteamDeck       bool            `json:"primarily_steam_deck"`
}

// AppReviewAuthor is the typed review author payload.
type AppReviewAuthor struct {
	SteamID              string `json:"steamid"`
	NumGamesOwned        int    `json:"num_games_owned"`
	NumReviews           int    `json:"num_reviews"`
	PlaytimeForever      int    `json:"playtime_forever"`
	PlaytimeLastTwoWeeks int    `json:"playtime_last_two_weeks"`
	PlaytimeAtReview     int    `json:"playtime_at_review"`
	LastPlayed           int64  `json:"last_played"`
}

// StorePlatforms is the stable platform support payload shared by Storefront methods.
type StorePlatforms struct {
	Windows bool `json:"windows"`
	Mac     bool `json:"mac"`
	Linux   bool `json:"linux"`
}

// StoreCategory is one Storefront category row.
type StoreCategory struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

// StoreGenre is one Storefront genre row.
type StoreGenre struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// StorePrice is one Storefront price overview payload.
type StorePrice struct {
	Currency         string `json:"currency"`
	Initial          int    `json:"initial"`
	Final            int    `json:"final"`
	DiscountPercent  int    `json:"discount_percent"`
	InitialFormatted string `json:"initial_formatted"`
	FinalFormatted   string `json:"final_formatted"`
}

// StorePackagePrice is one Storefront package price payload.
type StorePackagePrice struct {
	Currency        string `json:"currency"`
	Initial         int    `json:"initial"`
	Final           int    `json:"final"`
	DiscountPercent int    `json:"discount_percent"`
	Individual      int    `json:"individual"`
}

// StorePackageApp is one app row nested in a package response.
type StorePackageApp struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

// StoreReleaseDate is the stable release date payload.
type StoreReleaseDate struct {
	ComingSoon bool   `json:"coming_soon"`
	Date       string `json:"date"`
}

// FlexibleFloat64 accepts either a JSON number or a quoted numeric string.
type FlexibleFloat64 float64

// Float64 reports the underlying numeric value.
func (f FlexibleFloat64) Float64() float64 {
	return float64(f)
}

// UnmarshalJSON decodes one floating-point value from either a number or string payload.
func (f *FlexibleFloat64) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*f = 0
		return nil
	}

	if unquoted, err := strconv.Unquote(trimmed); err == nil {
		trimmed = strings.TrimSpace(unquoted)
	}

	value, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return err
	}
	*f = FlexibleFloat64(value)
	return nil
}
