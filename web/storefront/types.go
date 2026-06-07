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
	Type                string                   `json:"type"`
	Name                string                   `json:"name"`
	SteamAppID          uint32                   `json:"steam_appid"`
	RequiredAge         FlexibleInt              `json:"required_age"`
	IsFree              bool                     `json:"is_free"`
	ControllerSupport   string                   `json:"controller_support,omitempty"`
	DetailedDescription string                   `json:"detailed_description,omitempty"`
	AboutTheGame        string                   `json:"about_the_game,omitempty"`
	ShortDescription    string                   `json:"short_description"`
	SupportedLanguages  string                   `json:"supported_languages,omitempty"`
	HeaderImage         string                   `json:"header_image"`
	CapsuleImage        string                   `json:"capsule_image,omitempty"`
	CapsuleImageV5      string                   `json:"capsule_imagev5,omitempty"`
	Website             string                   `json:"website,omitempty"`
	PCRequirements      *StoreRequirements       `json:"pc_requirements,omitempty"`
	MacRequirements     *StoreRequirements       `json:"mac_requirements,omitempty"`
	LinuxRequirements   *StoreRequirements       `json:"linux_requirements,omitempty"`
	Developers          []string                 `json:"developers"`
	Publishers          []string                 `json:"publishers"`
	PriceOverview       *StorePrice              `json:"price_overview,omitempty"`
	Platforms           StorePlatforms           `json:"platforms"`
	Metacritic          *StoreMetacritic         `json:"metacritic,omitempty"`
	Categories          []StoreCategory          `json:"categories,omitempty"`
	Genres              []StoreGenre             `json:"genres,omitempty"`
	Screenshots         []StoreScreenshot        `json:"screenshots,omitempty"`
	Movies              []StoreMovie             `json:"movies,omitempty"`
	Recommendations     *StoreRecommendations    `json:"recommendations,omitempty"`
	Achievements        *StoreAchievements       `json:"achievements,omitempty"`
	Packages            []uint32                 `json:"packages,omitempty"`
	PackageGroups       json.RawMessage          `json:"package_groups,omitempty"`
	ReleaseDate         *StoreReleaseDate        `json:"release_date,omitempty"`
	SupportInfo         *StoreSupportInfo        `json:"support_info,omitempty"`
	Background          string                   `json:"background,omitempty"`
	BackgroundRaw       string                   `json:"background_raw,omitempty"`
	ContentDescriptors  *StoreContentDescriptors `json:"content_descriptors,omitempty"`
	Ratings             json.RawMessage          `json:"ratings,omitempty"`
}

// DecodeRatings decodes the app details ratings raw subtree into a typed map.
func (d AppDetailsData) DecodeRatings() (StoreRatings, error) {
	if len(d.Ratings) == 0 {
		return nil, nil
	}

	var ratings StoreRatings
	if err := json.Unmarshal(d.Ratings, &ratings); err != nil {
		return nil, err
	}
	return ratings, nil
}

// SteamGermanyRequiredAge returns ratings.steam_germany.required_age when present.
func (d AppDetailsData) SteamGermanyRequiredAge() (string, bool, error) {
	ratings, err := d.DecodeRatings()
	if err != nil {
		return "", false, err
	}
	rating, ok := ratings["steam_germany"]
	if !ok || rating.RequiredAge == "" {
		return "", false, nil
	}
	return rating.RequiredAge, true, nil
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

// StoreRequirements is one platform requirements block.
type StoreRequirements struct {
	Minimum     string `json:"minimum,omitempty"`
	Recommended string `json:"recommended,omitempty"`
}

// UnmarshalJSON accepts Steam's normal requirements object and its occasional
// empty array/string variants for apps without requirements.
func (r *StoreRequirements) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" || trimmed == "[]" || trimmed == `""` {
		*r = StoreRequirements{}
		return nil
	}

	type requirements StoreRequirements
	var decoded requirements
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*r = StoreRequirements(decoded)
	return nil
}

// StoreMetacritic is one Storefront metacritic payload.
type StoreMetacritic struct {
	Score int    `json:"score"`
	URL   string `json:"url"`
}

// StoreScreenshot is one Storefront screenshot payload.
type StoreScreenshot struct {
	ID            int    `json:"id"`
	PathThumbnail string `json:"path_thumbnail"`
	PathFull      string `json:"path_full"`
}

// StoreMovie is one Storefront movie/trailer payload.
type StoreMovie struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	Thumbnail string            `json:"thumbnail"`
	WebM      StoreMovieSources `json:"webm,omitempty"`
	MP4       StoreMovieSources `json:"mp4,omitempty"`
	DASHAV1   string            `json:"dash_av1,omitempty"`
	DASHH264  string            `json:"dash_h264,omitempty"`
	HLSH264   string            `json:"hls_h264,omitempty"`
	Highlight bool              `json:"highlight"`
}

// StoreMovieSources contains direct movie files when Steam includes them.
type StoreMovieSources struct {
	P480 string `json:"480,omitempty"`
	Max  string `json:"max,omitempty"`
}

// StoreRecommendations is the Storefront recommendation count payload.
type StoreRecommendations struct {
	Total int `json:"total"`
}

// StoreAchievements is the Storefront highlighted achievements payload.
type StoreAchievements struct {
	Total       int                `json:"total"`
	Highlighted []StoreAchievement `json:"highlighted,omitempty"`
}

// StoreAchievement is one highlighted achievement row.
type StoreAchievement struct {
	Icon          string `json:"icon"`
	LocalizedName string `json:"localized_name"`
	Name          string `json:"name"`
	Path          string `json:"path"`
}

// StoreSupportInfo is the Storefront support payload.
type StoreSupportInfo struct {
	URL   string `json:"url"`
	Email string `json:"email"`
}

// StoreContentDescriptors is the Storefront content descriptor payload.
type StoreContentDescriptors struct {
	IDs   []int  `json:"ids,omitempty"`
	Notes string `json:"notes,omitempty"`
}

// StoreRatings is the typed Storefront app ratings payload keyed by rating board.
type StoreRatings map[string]StoreRating

// StoreRating is one rating board payload. Raw preserves board-specific fields.
type StoreRating struct {
	Rating      string          `json:"rating,omitempty"`
	RequiredAge string          `json:"required_age,omitempty"`
	Raw         json.RawMessage `json:"-"`
}

// UnmarshalJSON stores each rating board's raw payload while decoding common fields.
func (r *StoreRating) UnmarshalJSON(data []byte) error {
	type storeRating StoreRating
	var decoded storeRating
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*r = StoreRating(decoded)
	r.Raw = append(r.Raw[:0], data...)
	return nil
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

// FlexibleInt accepts either a JSON integer or a quoted integer string.
type FlexibleInt int

// Int reports the underlying integer value.
func (f FlexibleInt) Int() int {
	return int(f)
}

// UnmarshalJSON decodes one integer value from either a number or string payload.
func (f *FlexibleInt) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*f = 0
		return nil
	}

	if unquoted, err := strconv.Unquote(trimmed); err == nil {
		trimmed = strings.TrimSpace(unquoted)
	}
	if trimmed == "" {
		*f = 0
		return nil
	}

	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return err
	}
	*f = FlexibleInt(value)
	return nil
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
