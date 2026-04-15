package steamnews

// GetNewsForAppResponse matches ISteamNews/GetNewsForApp/v2.
type GetNewsForAppResponse struct {
	AppNews AppNews `json:"appnews"`
}

// AppNews is the app news payload.
type AppNews struct {
	AppID     uint32     `json:"appid"`
	NewsItems []NewsItem `json:"newsitems"`
	Count     int        `json:"count"`
}

// NewsItem matches Valve's news item structure.
type NewsItem struct {
	GID           string `json:"gid"`
	Title         string `json:"title"`
	URL           string `json:"url"`
	IsExternalURL bool   `json:"is_external_url"`
	Author        string `json:"author"`
	Contents      string `json:"contents"`
	FeedLabel     string `json:"feedlabel"`
	Date          int64  `json:"date"`
	FeedName      string `json:"feedname"`
	FeedType      int    `json:"feed_type"`
	AppID         uint32 `json:"appid"`
	Tags          string `json:"tags"`
}
