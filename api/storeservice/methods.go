package storeservice

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofurry/steam-go/internal/endpoint"
	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	"github.com/gofurry/steam-go/internal/response"
)

// GetAppListOptions controls optional query parameters for GetAppList.
type GetAppListOptions struct {
	IfModifiedSince         *uint32
	HaveDescriptionLanguage string
	IncludeGames            *bool
	IncludeDLC              *bool
	IncludeSoftware         *bool
	IncludeVideos           *bool
	IncludeHardware         *bool
	LastAppID               *uint32
	MaxResults              *uint32
}

// GetUserGameInterestStateOptions controls optional query parameters for GetUserGameInterestState.
type GetUserGameInterestStateOptions struct {
	StoreAppID *uint32
	BetaAppID  *uint32
}

// GetAppList returns paginated Steam app metadata.
func (s *Service) GetAppList(ctx context.Context, opts *GetAppListOptions) (GetAppListResponse, error) {
	body, err := s.GetAppListRaw(ctx, opts)
	if err != nil {
		return GetAppListResponse{}, err
	}
	return response.DecodeJSON[GetAppListResponse](body)
}

// GetAppListRaw returns the raw JSON response body.
func (s *Service) GetAppListRaw(ctx context.Context, opts *GetAppListOptions) ([]byte, error) {
	query := url.Values{}
	if opts != nil {
		if opts.IfModifiedSince != nil {
			query.Set("if_modified_since", strconv.FormatUint(uint64(*opts.IfModifiedSince), 10))
		}
		if language := strings.TrimSpace(opts.HaveDescriptionLanguage); language != "" {
			query.Set("have_description_language", language)
		}
		if opts.IncludeGames != nil {
			query.Set("include_games", strconv.FormatBool(*opts.IncludeGames))
		}
		if opts.IncludeDLC != nil {
			query.Set("include_dlc", strconv.FormatBool(*opts.IncludeDLC))
		}
		if opts.IncludeSoftware != nil {
			query.Set("include_software", strconv.FormatBool(*opts.IncludeSoftware))
		}
		if opts.IncludeVideos != nil {
			query.Set("include_videos", strconv.FormatBool(*opts.IncludeVideos))
		}
		if opts.IncludeHardware != nil {
			query.Set("include_hardware", strconv.FormatBool(*opts.IncludeHardware))
		}
		if opts.LastAppID != nil {
			query.Set("last_appid", strconv.FormatUint(uint64(*opts.LastAppID), 10))
		}
		if opts.MaxResults != nil {
			if *opts.MaxResults > 50000 {
				return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max results cannot exceed 50000", nil, nil)
			}
			query.Set("max_results", strconv.FormatUint(uint64(*opts.MaxResults), 10))
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.StoreServiceGetAppList,
		Query:  query,
	})
}

// GetGamesFollowed returns app ids followed by a Steam user.
func (s *Service) GetGamesFollowed(ctx context.Context, steamID string) (GetGamesFollowedResponse, error) {
	body, err := s.GetGamesFollowedRaw(ctx, steamID)
	if err != nil {
		return GetGamesFollowedResponse{}, err
	}
	return response.DecodeJSON[GetGamesFollowedResponse](body)
}

// GetGamesFollowedRaw returns the raw JSON response body.
func (s *Service) GetGamesFollowedRaw(ctx context.Context, steamID string) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.StoreServiceGetGamesFollowed,
		Query:  query,
	})
}

// GetGamesFollowedCount returns followed game count for a Steam user.
func (s *Service) GetGamesFollowedCount(ctx context.Context, steamID string) (GetGamesFollowedCountResponse, error) {
	body, err := s.GetGamesFollowedCountRaw(ctx, steamID)
	if err != nil {
		return GetGamesFollowedCountResponse{}, err
	}
	return response.DecodeJSON[GetGamesFollowedCountResponse](body)
}

// GetGamesFollowedCountRaw returns the raw JSON response body.
func (s *Service) GetGamesFollowedCountRaw(ctx context.Context, steamID string) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.StoreServiceGetGamesFollowedCount,
		Query:  query,
	})
}

// GetMostPopularTags returns Steam's current popular tags.
func (s *Service) GetMostPopularTags(ctx context.Context) (GetMostPopularTagsResponse, error) {
	body, err := s.GetMostPopularTagsRaw(ctx)
	if err != nil {
		return GetMostPopularTagsResponse{}, err
	}
	return response.DecodeJSON[GetMostPopularTagsResponse](body)
}

// GetMostPopularTagsRaw returns the raw JSON response body.
func (s *Service) GetMostPopularTagsRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.StoreServiceGetMostPopularTags,
	})
}

// GetUserGameInterestState returns ownership/following or queue state for a user and app.
func (s *Service) GetUserGameInterestState(ctx context.Context, accessToken string, appID uint32, opts *GetUserGameInterestStateOptions) (GetUserGameInterestStateResponse, error) {
	body, err := s.GetUserGameInterestStateRaw(ctx, accessToken, appID, opts)
	if err != nil {
		return GetUserGameInterestStateResponse{}, err
	}
	return response.DecodeJSON[GetUserGameInterestStateResponse](body)
}

// GetUserGameInterestStateRaw returns the raw JSON response body.
func (s *Service) GetUserGameInterestStateRaw(ctx context.Context, accessToken string, appID uint32, opts *GetUserGameInterestStateOptions) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}

	hasIdentifier := false
	if appID > 0 {
		query.Set("appid", strconv.FormatUint(uint64(appID), 10))
		hasIdentifier = true
	}
	if opts != nil {
		if opts.StoreAppID != nil {
			query.Set("store_appid", strconv.FormatUint(uint64(*opts.StoreAppID), 10))
			hasIdentifier = true
		}
		if opts.BetaAppID != nil {
			query.Set("beta_appid", strconv.FormatUint(uint64(*opts.BetaAppID), 10))
			hasIdentifier = true
		}
	}
	if !hasIdentifier {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "at least one app identifier is required", nil, nil)
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method:    http.MethodPost,
		Path:      endpoint.StoreServiceGetUserGameInterestState,
		Query:     query,
		Retryable: request.Retryable(true),
	})
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

func buildSteamIDQuery(steamID string) (url.Values, error) {
	trimmed := strings.TrimSpace(steamID)
	if trimmed == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id is required", nil, nil)
	}

	query := url.Values{}
	query.Set("steamid", trimmed)
	return query, nil
}
