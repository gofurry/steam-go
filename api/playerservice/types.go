package playerservice

// GetOwnedGamesResponse matches IPlayerService/GetOwnedGames/v1.
type GetOwnedGamesResponse struct {
	Response OwnedGames `json:"response"`
}

// OwnedGames is the top-level owned games payload.
type OwnedGames struct {
	GameCount int         `json:"game_count"`
	Games     []OwnedGame `json:"games"`
}

// OwnedGame matches the Valve owned-game schema.
type OwnedGame struct {
	AppID                  uint32 `json:"appid"`
	Name                   string `json:"name"`
	Playtime2Weeks         int    `json:"playtime_2weeks"`
	PlaytimeForever        int    `json:"playtime_forever"`
	ImgIconURL             string `json:"img_icon_url"`
	HasCommunityVisible    bool   `json:"has_community_visible_stats"`
	PlaytimeWindowsForever int    `json:"playtime_windows_forever"`
	PlaytimeMacForever     int    `json:"playtime_mac_forever"`
	PlaytimeLinuxForever   int    `json:"playtime_linux_forever"`
	PlaytimeDeckForever    int    `json:"playtime_deck_forever"`
	RTimeLastPlayed        int64  `json:"rtime_last_played"`
	CapsuleFilename        string `json:"capsule_filename"`
	HasWorkshop            bool   `json:"has_workshop"`
	HasMarket              bool   `json:"has_market"`
	HasDLC                 bool   `json:"has_dlc"`
	ContentDescriptorIDs   []int  `json:"content_descriptorids"`
	PlaytimeDisconnected   int    `json:"playtime_disconnected"`
}

// ClientGetLastPlayedTimesOptions controls optional query parameters for ClientGetLastPlayedTimes.
type ClientGetLastPlayedTimesOptions struct {
	MinLastPlayed uint32
}

// ClientGetLastPlayedTimesResponse matches IPlayerService/ClientGetLastPlayedTimes/v1.
type ClientGetLastPlayedTimesResponse struct {
	Response LastPlayedTimes `json:"response"`
}

// LastPlayedTimes is the top-level last-played payload.
type LastPlayedTimes struct {
	Games []LastPlayedGame `json:"games"`
}

// LastPlayedGame matches one game entry in ClientGetLastPlayedTimes.
type LastPlayedGame struct {
	AppID                  uint32 `json:"appid"`
	LastPlaytime           int64  `json:"last_playtime"`
	Playtime2Weeks         int    `json:"playtime_2weeks"`
	PlaytimeForever        int    `json:"playtime_forever"`
	FirstPlaytime          int64  `json:"first_playtime"`
	PlaytimeWindowsForever int    `json:"playtime_windows_forever"`
	PlaytimeMacForever     int    `json:"playtime_mac_forever"`
	PlaytimeLinuxForever   int    `json:"playtime_linux_forever"`
	PlaytimeDeckForever    int    `json:"playtime_deck_forever"`
	FirstWindowsPlaytime   int64  `json:"first_windows_playtime"`
	FirstMacPlaytime       int64  `json:"first_mac_playtime"`
	FirstLinuxPlaytime     int64  `json:"first_linux_playtime"`
	FirstDeckPlaytime      int64  `json:"first_deck_playtime"`
	LastWindowsPlaytime    int64  `json:"last_windows_playtime"`
	LastMacPlaytime        int64  `json:"last_mac_playtime"`
	LastLinuxPlaytime      int64  `json:"last_linux_playtime"`
	LastDeckPlaytime       int64  `json:"last_deck_playtime"`
	PlaytimeDisconnected   int    `json:"playtime_disconnected"`
}

// GetAchievementsProgressOptions controls optional query parameters for GetAchievementsProgress.
type GetAchievementsProgressOptions struct {
	SteamID             string
	Language            string
	AppIDs              []uint32
	IncludeUnvettedApps *bool
}

// GetAchievementsProgressResponse matches IPlayerService/GetAchievementsProgress/v1.
type GetAchievementsProgressResponse struct {
	Response AchievementsProgressPayload `json:"response"`
}

// AchievementsProgressPayload is the top-level achievements-progress payload.
type AchievementsProgressPayload struct {
	AchievementProgress []AchievementProgressEntry `json:"achievement_progress"`
}

// AchievementProgressEntry matches one app entry in GetAchievementsProgress.
type AchievementProgressEntry struct {
	AppID       uint32  `json:"appid"`
	Unlocked    int     `json:"unlocked"`
	Total       int     `json:"total"`
	Percentage  float64 `json:"percentage"`
	AllUnlocked bool    `json:"all_unlocked"`
	CacheTime   int64   `json:"cache_time"`
	Vetted      bool    `json:"vetted"`
}

// PlayerProfileItemOptions controls optional query parameters for profile-item lookups.
type PlayerProfileItemOptions struct {
	Language string
}

// GetAnimatedAvatarResponse matches IPlayerService/GetAnimatedAvatar/v1.
type GetAnimatedAvatarResponse struct {
	Response AnimatedAvatarPayload `json:"response"`
}

// AnimatedAvatarPayload is the top-level animated-avatar payload.
type AnimatedAvatarPayload struct {
	Avatar ProfileItem `json:"avatar"`
}

// GetAvatarFrameResponse matches IPlayerService/GetAvatarFrame/v1.
type GetAvatarFrameResponse struct {
	Response AvatarFramePayload `json:"response"`
}

// AvatarFramePayload is the top-level avatar-frame payload.
type AvatarFramePayload struct {
	AvatarFrame ProfileItem `json:"avatar_frame"`
}

// ProfileItem matches Steam profile customization assets such as avatars and frames.
type ProfileItem struct {
	CommunityItemID string         `json:"communityitemid"`
	ImageSmall      string         `json:"image_small"`
	ImageLarge      string         `json:"image_large"`
	Name            string         `json:"name"`
	ItemTitle       string         `json:"item_title"`
	ItemDescription string         `json:"item_description"`
	AppID           uint32         `json:"appid"`
	ItemType        int            `json:"item_type"`
	ItemClass       int            `json:"item_class"`
	MovieWebM       string         `json:"movie_webm"`
	MovieMP4        string         `json:"movie_mp4"`
	MovieWebMSmall  string         `json:"movie_webm_small"`
	MovieMP4Small   string         `json:"movie_mp4_small"`
	ProfileColors   []ProfileColor `json:"profile_colors"`
}

// ProfileColor matches one style color override attached to a profile item.
type ProfileColor struct {
	StyleName string `json:"style_name"`
	Color     string `json:"color"`
}

// GetMiniProfileBackgroundResponse matches IPlayerService/GetMiniProfileBackground/v1.
type GetMiniProfileBackgroundResponse struct {
	Response MiniProfileBackgroundPayload `json:"response"`
}

// MiniProfileBackgroundPayload is the top-level mini-profile background payload.
type MiniProfileBackgroundPayload struct {
	ProfileBackground ProfileItem `json:"profile_background"`
}

// GetNicknameListResponse matches IPlayerService/GetNicknameList/v1.
type GetNicknameListResponse struct {
	Response NicknameListPayload `json:"response"`
}

// NicknameListPayload is the top-level nickname list payload.
type NicknameListPayload struct {
	Nicknames []NicknameEntry `json:"nicknames"`
}

// NicknameEntry matches one stored nickname record.
type NicknameEntry struct {
	AccountID uint32 `json:"accountid"`
	Nickname  string `json:"nickname"`
}

// GetPlayerLinkDetailsResponse matches IPlayerService/GetPlayerLinkDetails/v1.
type GetPlayerLinkDetailsResponse struct {
	Response PlayerLinkDetailsPayload `json:"response"`
}

// PlayerLinkDetailsPayload is the top-level player-link details payload.
type PlayerLinkDetailsPayload struct {
	Accounts []PlayerLinkDetailsAccount `json:"accounts"`
}

// PlayerLinkDetailsAccount groups public and private details for a linked player.
type PlayerLinkDetailsAccount struct {
	PublicData  PlayerLinkDetailsPublicData  `json:"public_data"`
	PrivateData PlayerLinkDetailsPrivateData `json:"private_data"`
}

// PlayerLinkDetailsPublicData matches public-facing link details.
type PlayerLinkDetailsPublicData struct {
	SteamID                  string `json:"steamid"`
	VisibilityState          int    `json:"visibility_state"`
	ProfileState             int    `json:"profile_state"`
	SHADigestAvatar          string `json:"sha_digest_avatar"`
	PersonaName              string `json:"persona_name"`
	ProfileURL               string `json:"profile_url"`
	ContentCountryRestricted bool   `json:"content_country_restricted"`
}

// PlayerLinkDetailsPrivateData matches private timing details exposed by the endpoint.
type PlayerLinkDetailsPrivateData struct {
	TimeCreated    int64 `json:"time_created"`
	LastLogoffTime int64 `json:"last_logoff_time"`
	LastSeenOnline int64 `json:"last_seen_online"`
}

// GetProfileBackgroundResponse matches IPlayerService/GetProfileBackground/v1.
type GetProfileBackgroundResponse struct {
	Response ProfileBackgroundPayload `json:"response"`
}

// ProfileBackgroundPayload is the top-level profile background payload.
type ProfileBackgroundPayload struct {
	ProfileBackground ProfileItem `json:"profile_background"`
}

// GetProfileCustomizationResponse matches IPlayerService/GetProfileCustomization/v1.
type GetProfileCustomizationResponse struct {
	Response ProfileCustomizationPayload `json:"response"`
}

// ProfileCustomizationPayload is the top-level profile customization payload.
type ProfileCustomizationPayload struct {
	Customizations          []ProfileCustomization          `json:"customizations"`
	SlotsAvailable          int                             `json:"slots_available"`
	ProfileTheme            ProfileTheme                    `json:"profile_theme"`
	PurchasedCustomizations []PurchasedProfileCustomization `json:"purchased_customizations"`
	ProfilePreferences      ProfilePreferences              `json:"profile_preferences"`
}

// ProfileCustomization matches one customization module on a profile.
type ProfileCustomization struct {
	CustomizationType  int                        `json:"customization_type"`
	Large              bool                       `json:"large"`
	Slots              []ProfileCustomizationSlot `json:"slots"`
	Active             bool                       `json:"active"`
	CustomizationStyle int                        `json:"customization_style"`
	PurchaseID         string                     `json:"purchaseid"`
	Level              int                        `json:"level"`
}

// ProfileCustomizationSlot is a flexible slot payload used across many customization modules.
type ProfileCustomizationSlot struct {
	Slot            int    `json:"slot"`
	AppID           uint32 `json:"appid"`
	Title           string `json:"title"`
	ReplayYear      int    `json:"replay_year"`
	AccountID       uint32 `json:"accountid"`
	BadgeID         int    `json:"badgeid"`
	Notes           string `json:"notes"`
	BanCheckResult  int    `json:"ban_check_result"`
	PublishedFileID string `json:"publishedfileid"`
	ItemAssetID     string `json:"item_assetid"`
	ItemContextID   string `json:"item_contextid"`
	ItemClassID     string `json:"item_classid"`
	ItemInstanceID  string `json:"item_instanceid"`
}

// PurchasedProfileCustomization matches one purchased customization upgrade.
type PurchasedProfileCustomization struct {
	PurchaseID        string `json:"purchaseid"`
	CustomizationType int    `json:"customization_type"`
	Level             int    `json:"level"`
}

// ProfileTheme matches one profile theme definition.
type ProfileTheme struct {
	ThemeID string `json:"theme_id"`
	Title   string `json:"title"`
}

// ProfilePreferences matches profile-level presentation preferences.
type ProfilePreferences struct {
	HideProfileAwards bool `json:"hide_profile_awards"`
}

// GetProfileItemsEquippedResponse matches IPlayerService/GetProfileItemsEquipped/v1.
type GetProfileItemsEquippedResponse struct {
	Response ProfileItemsEquippedPayload `json:"response"`
}

// ProfileItemsEquippedPayload is the top-level equipped profile-items payload.
type ProfileItemsEquippedPayload struct {
	ProfileBackground     ProfileItem `json:"profile_background"`
	MiniProfileBackground ProfileItem `json:"mini_profile_background"`
	AvatarFrame           ProfileItem `json:"avatar_frame"`
	AnimatedAvatar        ProfileItem `json:"animated_avatar"`
	ProfileModifier       ProfileItem `json:"profile_modifier"`
	SteamDeckKeyboardSkin ProfileItem `json:"steam_deck_keyboard_skin"`
}

// GetProfileItemsOwnedResponse matches IPlayerService/GetProfileItemsOwned/v1.
type GetProfileItemsOwnedResponse struct {
	Response ProfileItemsOwnedPayload `json:"response"`
}

// ProfileItemsOwnedPayload is the top-level owned profile-items payload.
type ProfileItemsOwnedPayload struct {
	ProfileBackgrounds     []ProfileItem `json:"profile_backgrounds"`
	MiniProfileBackgrounds []ProfileItem `json:"mini_profile_backgrounds"`
	AvatarFrames           []ProfileItem `json:"avatar_frames"`
	AnimatedAvatars        []ProfileItem `json:"animated_avatars"`
	ProfileModifiers       []ProfileItem `json:"profile_modifiers"`
	SteamDeckKeyboardSkins []ProfileItem `json:"steam_deck_keyboard_skins"`
}

// GetProfileThemesAvailableResponse matches IPlayerService/GetProfileThemesAvailable/v1.
type GetProfileThemesAvailableResponse struct {
	Response ProfileThemesAvailablePayload `json:"response"`
}

// ProfileThemesAvailablePayload is the top-level available profile-themes payload.
type ProfileThemesAvailablePayload struct {
	ProfileThemes []ProfileTheme `json:"profile_themes"`
}

// GetBadgesResponse matches IPlayerService/GetBadges/v1.
type GetBadgesResponse struct {
	Response BadgesPayload `json:"response"`
}

// BadgesPayload is the top-level badges payload.
type BadgesPayload struct {
	Badges                     []Badge `json:"badges"`
	PlayerXP                   int     `json:"player_xp"`
	PlayerLevel                int     `json:"player_level"`
	PlayerXPNeededToLevelUp    int     `json:"player_xp_needed_to_level_up"`
	PlayerXPNeededCurrentLevel int     `json:"player_xp_needed_current_level"`
}

// Badge matches one badge entry returned by GetBadges.
type Badge struct {
	BadgeID         int    `json:"badgeid"`
	AppID           uint32 `json:"appid"`
	Level           int    `json:"level"`
	CompletionTime  int64  `json:"completion_time"`
	XP              int    `json:"xp"`
	CommunityItemID string `json:"communityitemid"`
	BorderColor     int    `json:"border_color"`
	Scarcity        int    `json:"scarcity"`
}

// GetCommunityBadgeProgressResponse matches IPlayerService/GetCommunityBadgeProgress/v1.
type GetCommunityBadgeProgressResponse struct {
	Response CommunityBadgeProgressPayload `json:"response"`
}

// CommunityBadgeProgressPayload is the top-level community-badge payload.
type CommunityBadgeProgressPayload struct {
	Quests []CommunityQuestProgress `json:"quests"`
}

// CommunityQuestProgress matches one community quest flag.
type CommunityQuestProgress struct {
	QuestID   int  `json:"questid"`
	Completed bool `json:"completed"`
}

// GetCommunityPreferencesResponse matches IPlayerService/GetCommunityPreferences/v1.
type GetCommunityPreferencesResponse struct {
	Response CommunityPreferencesPayload `json:"response"`
}

// CommunityPreferencesPayload is the top-level community-preferences payload.
type CommunityPreferencesPayload struct {
	Preferences                  CommunityPreferences `json:"preferences"`
	ContentDescriptorPreferences map[string]any       `json:"content_descriptor_preferences"`
}

// CommunityPreferences matches text-filter and nickname preference settings.
type CommunityPreferences struct {
	ParenthesizeNicknames   bool  `json:"parenthesize_nicknames"`
	TextFilterSetting       int   `json:"text_filter_setting"`
	TextFilterIgnoreFriends bool  `json:"text_filter_ignore_friends"`
	TextFilterWordsRevision int   `json:"text_filter_words_revision"`
	TimestampUpdated        int64 `json:"timestamp_updated"`
}

// GetFavoriteBadgeResponse matches IPlayerService/GetFavoriteBadge/v1.
type GetFavoriteBadgeResponse struct {
	Response FavoriteBadgePayload `json:"response"`
}

// FavoriteBadgePayload is the top-level favorite-badge payload.
type FavoriteBadgePayload struct {
	HasFavoriteBadge bool   `json:"has_favorite_badge"`
	CommunityItemID  string `json:"communityitemid"`
	ItemType         int    `json:"item_type"`
	BorderColor      int    `json:"border_color"`
	AppID            uint32 `json:"appid"`
	Level            int    `json:"level"`
}

// GetFriendsGameplayInfoResponse matches IPlayerService/GetFriendsGameplayInfo/v1.
type GetFriendsGameplayInfoResponse struct {
	Response FriendsGameplayInfoPayload `json:"response"`
}

// FriendsGameplayInfoPayload is the top-level GetFriendsGameplayInfo payload.
type FriendsGameplayInfoPayload struct {
	YourInfo       FriendsGameplayInfoYourInfo `json:"your_info"`
	PlayedRecently []FriendsGameplayInfoPeer   `json:"played_recently"`
	PlayedEver     []FriendsGameplayInfoPeer   `json:"played_ever"`
	Owns           []FriendsGameplayInfoOwner  `json:"owns"`
	InWishlist     []FriendsGameplayInfoOwner  `json:"in_wishlist"`
}

// FriendsGameplayInfoYourInfo matches the caller summary in GetFriendsGameplayInfo.
type FriendsGameplayInfoYourInfo struct {
	SteamID              string `json:"steamid"`
	MinutesPlayed        int    `json:"minutes_played"`
	MinutesPlayedForever int    `json:"minutes_played_forever"`
	Owned                bool   `json:"owned"`
}

// FriendsGameplayInfoPeer matches one related player entry in GetFriendsGameplayInfo.
type FriendsGameplayInfoPeer struct {
	SteamID              string `json:"steamid"`
	MinutesPlayed        int    `json:"minutes_played"`
	MinutesPlayedForever int    `json:"minutes_played_forever"`
}

// FriendsGameplayInfoOwner matches simple steamid-only player entries.
type FriendsGameplayInfoOwner struct {
	SteamID string `json:"steamid"`
}
