package playerservice

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/GoFurry/steam-go/internal/endpoint"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/response"
)

// GetOwnedGamesOptions controls optional query parameters for GetOwnedGames.
type GetOwnedGamesOptions struct {
	IncludePlayedFreeGames bool
	AppIDsFilter           []uint32
}

// GetOwnedGames returns the typed owned games payload.
func (s *Service) GetOwnedGames(ctx context.Context, steamID string, opts *GetOwnedGamesOptions) (GetOwnedGamesResponse, error) {
	body, err := s.GetOwnedGamesRaw(ctx, steamID, opts)
	if err != nil {
		return GetOwnedGamesResponse{}, err
	}
	return response.DecodeJSON[GetOwnedGamesResponse](body)
}

// GetOwnedGamesRaw returns the raw JSON response body.
func (s *Service) GetOwnedGamesRaw(ctx context.Context, steamID string, opts *GetOwnedGamesOptions) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}
	query.Set("include_appinfo", "1")
	if opts != nil {
		if opts.IncludePlayedFreeGames {
			query.Set("include_played_free_games", "1")
		}
		for _, appID := range opts.AppIDsFilter {
			query.Add("appids_filter", strconv.FormatUint(uint64(appID), 10))
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetOwnedGames,
		Query:  query,
	})
}

// ClientGetLastPlayedTimes returns recent playtime details for the authenticated caller.
func (s *Service) ClientGetLastPlayedTimes(ctx context.Context, accessToken string, opts *ClientGetLastPlayedTimesOptions) (ClientGetLastPlayedTimesResponse, error) {
	body, err := s.ClientGetLastPlayedTimesRaw(ctx, accessToken, opts)
	if err != nil {
		return ClientGetLastPlayedTimesResponse{}, err
	}
	return response.DecodeJSON[ClientGetLastPlayedTimesResponse](body)
}

// ClientGetLastPlayedTimesRaw returns the raw JSON response body.
func (s *Service) ClientGetLastPlayedTimesRaw(ctx context.Context, accessToken string, opts *ClientGetLastPlayedTimesOptions) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}
	if opts != nil && opts.MinLastPlayed > 0 {
		query.Set("min_last_played", strconv.FormatUint(uint64(opts.MinLastPlayed), 10))
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceClientGetLastPlayedTimes,
		Query:  query,
	})
}

// GetAchievementsProgress returns aggregated achievement progress for one or more apps.
func (s *Service) GetAchievementsProgress(ctx context.Context, accessToken string, opts *GetAchievementsProgressOptions) (GetAchievementsProgressResponse, error) {
	body, err := s.GetAchievementsProgressRaw(ctx, accessToken, opts)
	if err != nil {
		return GetAchievementsProgressResponse{}, err
	}
	return response.DecodeJSON[GetAchievementsProgressResponse](body)
}

// GetAchievementsProgressRaw returns the raw JSON response body.
func (s *Service) GetAchievementsProgressRaw(ctx context.Context, accessToken string, opts *GetAchievementsProgressOptions) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		if strings.TrimSpace(opts.SteamID) != "" {
			query.Set("steamid", strings.TrimSpace(opts.SteamID))
		}
		if strings.TrimSpace(opts.Language) != "" {
			query.Set("language", strings.TrimSpace(opts.Language))
		}
		for idx, appID := range opts.AppIDs {
			if appID == 0 {
				return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
			}
			query.Set("appids["+strconv.Itoa(idx)+"]", strconv.FormatUint(uint64(appID), 10))
		}
		if opts.IncludeUnvettedApps != nil {
			query.Set("include_unvetted_apps", boolString(*opts.IncludeUnvettedApps))
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodPost,
		Path:   endpoint.PlayerServiceGetAchievementsProgress,
		Query:  query,
	})
}

// GetAnimatedAvatar returns the animated avatar item equipped by the given player.
func (s *Service) GetAnimatedAvatar(ctx context.Context, steamID string, opts *PlayerProfileItemOptions) (GetAnimatedAvatarResponse, error) {
	body, err := s.GetAnimatedAvatarRaw(ctx, steamID, opts)
	if err != nil {
		return GetAnimatedAvatarResponse{}, err
	}
	return response.DecodeJSON[GetAnimatedAvatarResponse](body)
}

// GetAnimatedAvatarRaw returns the raw JSON response body.
func (s *Service) GetAnimatedAvatarRaw(ctx context.Context, steamID string, opts *PlayerProfileItemOptions) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}
	applyLanguage(query, opts)

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetAnimatedAvatar,
		Query:  query,
	})
}

// GetAvatarFrame returns the avatar frame item equipped by the given player.
func (s *Service) GetAvatarFrame(ctx context.Context, steamID string, opts *PlayerProfileItemOptions) (GetAvatarFrameResponse, error) {
	body, err := s.GetAvatarFrameRaw(ctx, steamID, opts)
	if err != nil {
		return GetAvatarFrameResponse{}, err
	}
	return response.DecodeJSON[GetAvatarFrameResponse](body)
}

// GetAvatarFrameRaw returns the raw JSON response body.
func (s *Service) GetAvatarFrameRaw(ctx context.Context, steamID string, opts *PlayerProfileItemOptions) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}
	applyLanguage(query, opts)

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetAvatarFrame,
		Query:  query,
	})
}

// GetMiniProfileBackground returns the mini-profile background item equipped by the given player.
func (s *Service) GetMiniProfileBackground(ctx context.Context, steamID string, opts *PlayerProfileItemOptions) (GetMiniProfileBackgroundResponse, error) {
	body, err := s.GetMiniProfileBackgroundRaw(ctx, steamID, opts)
	if err != nil {
		return GetMiniProfileBackgroundResponse{}, err
	}
	return response.DecodeJSON[GetMiniProfileBackgroundResponse](body)
}

// GetMiniProfileBackgroundRaw returns the raw JSON response body.
func (s *Service) GetMiniProfileBackgroundRaw(ctx context.Context, steamID string, opts *PlayerProfileItemOptions) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}
	applyLanguage(query, opts)

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetMiniProfileBackground,
		Query:  query,
	})
}

// GetBadges returns badge progress and XP totals for the given player.
func (s *Service) GetBadges(ctx context.Context, steamID string) (GetBadgesResponse, error) {
	body, err := s.GetBadgesRaw(ctx, steamID)
	if err != nil {
		return GetBadgesResponse{}, err
	}
	return response.DecodeJSON[GetBadgesResponse](body)
}

// GetBadgesRaw returns the raw JSON response body.
func (s *Service) GetBadgesRaw(ctx context.Context, steamID string) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetBadges,
		Query:  query,
	})
}

// GetCommunityBadgeProgress returns quest completion data for the community badge.
func (s *Service) GetCommunityBadgeProgress(ctx context.Context, steamID string) (GetCommunityBadgeProgressResponse, error) {
	body, err := s.GetCommunityBadgeProgressRaw(ctx, steamID)
	if err != nil {
		return GetCommunityBadgeProgressResponse{}, err
	}
	return response.DecodeJSON[GetCommunityBadgeProgressResponse](body)
}

// GetCommunityBadgeProgressRaw returns the raw JSON response body.
func (s *Service) GetCommunityBadgeProgressRaw(ctx context.Context, steamID string) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetCommunityBadgeProgress,
		Query:  query,
	})
}

// GetNicknameList returns the caller's stored friend nickname list.
func (s *Service) GetNicknameList(ctx context.Context, accessToken string) (GetNicknameListResponse, error) {
	body, err := s.GetNicknameListRaw(ctx, accessToken)
	if err != nil {
		return GetNicknameListResponse{}, err
	}
	return response.DecodeJSON[GetNicknameListResponse](body)
}

// GetNicknameListRaw returns the raw JSON response body.
func (s *Service) GetNicknameListRaw(ctx context.Context, accessToken string) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetNicknameList,
		Query:  query,
	})
}

// GetPlayerLinkDetails returns profile link data for one or more Steam IDs.
func (s *Service) GetPlayerLinkDetails(ctx context.Context, steamIDs []string) (GetPlayerLinkDetailsResponse, error) {
	body, err := s.GetPlayerLinkDetailsRaw(ctx, steamIDs)
	if err != nil {
		return GetPlayerLinkDetailsResponse{}, err
	}
	return response.DecodeJSON[GetPlayerLinkDetailsResponse](body)
}

// GetPlayerLinkDetailsRaw returns the raw JSON response body.
func (s *Service) GetPlayerLinkDetailsRaw(ctx context.Context, steamIDs []string) ([]byte, error) {
	query, err := buildSteamIDsQuery(steamIDs)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetPlayerLinkDetails,
		Query:  query,
	})
}

// GetProfileBackground returns the profile background item equipped by the given player.
func (s *Service) GetProfileBackground(ctx context.Context, steamID string, opts *PlayerProfileItemOptions) (GetProfileBackgroundResponse, error) {
	body, err := s.GetProfileBackgroundRaw(ctx, steamID, opts)
	if err != nil {
		return GetProfileBackgroundResponse{}, err
	}
	return response.DecodeJSON[GetProfileBackgroundResponse](body)
}

// GetProfileBackgroundRaw returns the raw JSON response body.
func (s *Service) GetProfileBackgroundRaw(ctx context.Context, steamID string, opts *PlayerProfileItemOptions) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}
	applyLanguage(query, opts)

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetProfileBackground,
		Query:  query,
	})
}

// GetProfileCustomizationOptions controls optional query parameters for GetProfileCustomization.
type GetProfileCustomizationOptions struct {
	IncludeInactiveCustomizations  bool
	IncludePurchasedCustomizations bool
}

// GetProfileCustomization returns profile customization modules and related metadata.
func (s *Service) GetProfileCustomization(ctx context.Context, steamID string, opts *GetProfileCustomizationOptions) (GetProfileCustomizationResponse, error) {
	body, err := s.GetProfileCustomizationRaw(ctx, steamID, opts)
	if err != nil {
		return GetProfileCustomizationResponse{}, err
	}
	return response.DecodeJSON[GetProfileCustomizationResponse](body)
}

// GetProfileCustomizationRaw returns the raw JSON response body.
func (s *Service) GetProfileCustomizationRaw(ctx context.Context, steamID string, opts *GetProfileCustomizationOptions) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		if opts.IncludeInactiveCustomizations {
			query.Set("include_inactive_customizations", "true")
		}
		if opts.IncludePurchasedCustomizations {
			query.Set("include_purchased_customizations", "true")
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetProfileCustomization,
		Query:  query,
	})
}

// GetProfileItemsEquipped returns the profile items currently equipped by the given player.
func (s *Service) GetProfileItemsEquipped(ctx context.Context, steamID string, opts *PlayerProfileItemOptions) (GetProfileItemsEquippedResponse, error) {
	body, err := s.GetProfileItemsEquippedRaw(ctx, steamID, opts)
	if err != nil {
		return GetProfileItemsEquippedResponse{}, err
	}
	return response.DecodeJSON[GetProfileItemsEquippedResponse](body)
}

// GetProfileItemsEquippedRaw returns the raw JSON response body.
func (s *Service) GetProfileItemsEquippedRaw(ctx context.Context, steamID string, opts *PlayerProfileItemOptions) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}
	applyLanguage(query, opts)

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetProfileItemsEquipped,
		Query:  query,
	})
}

// GetProfileItemsOwnedOptions controls optional query parameters for GetProfileItemsOwned.
type GetProfileItemsOwnedOptions struct {
	Language string
	Filters  []int32
}

// GetProfileItemsOwned returns the caller's owned profile item inventory by item class.
func (s *Service) GetProfileItemsOwned(ctx context.Context, accessToken string, opts *GetProfileItemsOwnedOptions) (GetProfileItemsOwnedResponse, error) {
	body, err := s.GetProfileItemsOwnedRaw(ctx, accessToken, opts)
	if err != nil {
		return GetProfileItemsOwnedResponse{}, err
	}
	return response.DecodeJSON[GetProfileItemsOwnedResponse](body)
}

// GetProfileItemsOwnedRaw returns the raw JSON response body.
func (s *Service) GetProfileItemsOwnedRaw(ctx context.Context, accessToken string, opts *GetProfileItemsOwnedOptions) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		if strings.TrimSpace(opts.Language) != "" {
			query.Set("language", strings.TrimSpace(opts.Language))
		}
		for idx, filter := range opts.Filters {
			query.Set("filters["+strconv.Itoa(idx)+"]", strconv.FormatInt(int64(filter), 10))
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetProfileItemsOwned,
		Query:  query,
	})
}

// GetProfileThemesAvailable returns the caller's available profile themes.
func (s *Service) GetProfileThemesAvailable(ctx context.Context, accessToken string) (GetProfileThemesAvailableResponse, error) {
	body, err := s.GetProfileThemesAvailableRaw(ctx, accessToken)
	if err != nil {
		return GetProfileThemesAvailableResponse{}, err
	}
	return response.DecodeJSON[GetProfileThemesAvailableResponse](body)
}

// GetProfileThemesAvailableRaw returns the raw JSON response body.
func (s *Service) GetProfileThemesAvailableRaw(ctx context.Context, accessToken string) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetProfileThemesAvailable,
		Query:  query,
	})
}

// GetCommunityPreferences returns the caller's community and text-filter settings.
func (s *Service) GetCommunityPreferences(ctx context.Context, accessToken string) (GetCommunityPreferencesResponse, error) {
	body, err := s.GetCommunityPreferencesRaw(ctx, accessToken)
	if err != nil {
		return GetCommunityPreferencesResponse{}, err
	}
	return response.DecodeJSON[GetCommunityPreferencesResponse](body)
}

// GetCommunityPreferencesRaw returns the raw JSON response body.
func (s *Service) GetCommunityPreferencesRaw(ctx context.Context, accessToken string) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodPost,
		Path:   endpoint.PlayerServiceGetCommunityPreferences,
		Query:  query,
	})
}

// GetFavoriteBadge returns the badge the player chose to feature on their profile.
func (s *Service) GetFavoriteBadge(ctx context.Context, steamID string) (GetFavoriteBadgeResponse, error) {
	body, err := s.GetFavoriteBadgeRaw(ctx, steamID)
	if err != nil {
		return GetFavoriteBadgeResponse{}, err
	}
	return response.DecodeJSON[GetFavoriteBadgeResponse](body)
}

// GetFavoriteBadgeRaw returns the raw JSON response body.
func (s *Service) GetFavoriteBadgeRaw(ctx context.Context, steamID string) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetFavoriteBadge,
		Query:  query,
	})
}

// GetFriendsGameplayInfo returns friend gameplay relationship data for one app.
func (s *Service) GetFriendsGameplayInfo(ctx context.Context, accessToken string, appID uint32) (GetFriendsGameplayInfoResponse, error) {
	body, err := s.GetFriendsGameplayInfoRaw(ctx, accessToken, appID)
	if err != nil {
		return GetFriendsGameplayInfoResponse{}, err
	}
	return response.DecodeJSON[GetFriendsGameplayInfoResponse](body)
}

// GetFriendsGameplayInfoRaw returns the raw JSON response body.
func (s *Service) GetFriendsGameplayInfoRaw(ctx context.Context, accessToken string, appID uint32) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetFriendsGameplayInfo,
		Query:  query,
	})
}

func buildSteamIDQuery(steamID string) (url.Values, error) {
	trimmed := strings.TrimSpace(steamID)
	if trimmed == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id is required", nil, nil)
	}

	query := url.Values{}
	query.Set("steamid", trimmed)
	return query, nil
}

func buildSteamIDsQuery(steamIDs []string) (url.Values, error) {
	query := url.Values{}
	count := 0
	for idx, steamID := range steamIDs {
		trimmed := strings.TrimSpace(steamID)
		if trimmed == "" {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id is required", nil, nil)
		}
		query.Set("steamids["+strconv.Itoa(idx)+"]", trimmed)
		count++
	}
	if count == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "at least one steam id is required", nil, nil)
	}
	return query, nil
}

func buildAccessTokenQuery(accessToken string) (url.Values, error) {
	trimmed := strings.TrimSpace(accessToken)
	if trimmed == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "access token is required", nil, nil)
	}

	query := url.Values{}
	query.Set("access_token", trimmed)
	return query, nil
}

func applyLanguage(query url.Values, opts *PlayerProfileItemOptions) {
	if opts == nil {
		return
	}
	language := strings.TrimSpace(opts.Language)
	if language != "" {
		query.Set("language", language)
	}
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
