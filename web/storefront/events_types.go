package storefront

import "encoding/json"

// AdjacentPartnerEventsResponse is the Storefront adjacent partner events payload.
type AdjacentPartnerEventsResponse struct {
	Success int            `json:"success,omitempty"`
	Events  []PartnerEvent `json:"events"`
}

// PartnerEvent is the stable typed subset of one Storefront partner event.
type PartnerEvent struct {
	GID               string                  `json:"gid,omitempty"`
	ClanSteamID       string                  `json:"clan_steamid,omitempty"`
	AppID             uint32                  `json:"appid,omitempty"`
	EventName         string                  `json:"event_name,omitempty"`
	EventType         int                     `json:"event_type,omitempty"`
	CommentCount      int                     `json:"comment_count,omitempty"`
	ForumTopicID      string                  `json:"forum_topic_id,omitempty"`
	RTimeCreated      int64                   `json:"rtime_created,omitempty"`
	RTimeStart        int64                   `json:"rtime32_start_time,omitempty"`
	RTimeEnd          int64                   `json:"rtime32_end_time,omitempty"`
	RTimeLastModified int64                   `json:"rtime32_last_modified,omitempty"`
	Published         int                     `json:"published,omitempty"`
	AnnouncementBody  PartnerAnnouncementBody `json:"announcement_body"`
	Raw               json.RawMessage         `json:"-"`
}

// UnmarshalJSON stores the raw event payload while decoding the stable subset.
func (e *PartnerEvent) UnmarshalJSON(data []byte) error {
	type partnerEvent PartnerEvent
	var decoded partnerEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*e = PartnerEvent(decoded)
	e.Raw = append(e.Raw[:0], data...)
	return nil
}

// PartnerAnnouncementBody is the stable typed subset of an event announcement.
type PartnerAnnouncementBody struct {
	GID           string          `json:"gid,omitempty"`
	EventGID      string          `json:"event_gid,omitempty"`
	Headline      string          `json:"headline,omitempty"`
	Body          string          `json:"body,omitempty"`
	PostTime      int64           `json:"posttime,omitempty"`
	UpdateTime    int64           `json:"updatetime,omitempty"`
	URL           string          `json:"url,omitempty"`
	CommentCount  int             `json:"commentcount,omitempty"`
	ClanID        string          `json:"clanid,omitempty"`
	ForumTopicID  string          `json:"forum_topic_id,omitempty"`
	Tags          []string        `json:"tags,omitempty"`
	Language      int             `json:"language,omitempty"`
	VoteUpCount   int             `json:"voteupcount,omitempty"`
	VoteDownCount int             `json:"votedowncount,omitempty"`
	Raw           json.RawMessage `json:"-"`
}

// UnmarshalJSON stores the raw announcement payload while decoding the stable subset.
func (b *PartnerAnnouncementBody) UnmarshalJSON(data []byte) error {
	type partnerAnnouncementBody PartnerAnnouncementBody
	var decoded partnerAnnouncementBody
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*b = PartnerAnnouncementBody(decoded)
	b.Raw = append(b.Raw[:0], data...)
	return nil
}
