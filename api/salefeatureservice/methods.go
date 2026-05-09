package salefeatureservice

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

// GetFriendsSharedYearInReviewOptions controls optional query parameters for GetFriendsSharedYearInReview.
type GetFriendsSharedYearInReviewOptions struct {
	ReturnPrivate *bool
}

// GetFriendsSharedYearInReview returns friend Year in Review sharing states for the given player and year.
func (s *Service) GetFriendsSharedYearInReview(ctx context.Context, steamID string, year uint32, opts *GetFriendsSharedYearInReviewOptions) (GetFriendsSharedYearInReviewResponse, error) {
	body, err := s.GetFriendsSharedYearInReviewRaw(ctx, steamID, year, opts)
	if err != nil {
		return GetFriendsSharedYearInReviewResponse{}, err
	}
	return response.DecodeJSON[GetFriendsSharedYearInReviewResponse](body)
}

// GetFriendsSharedYearInReviewRaw returns the raw JSON response body.
func (s *Service) GetFriendsSharedYearInReviewRaw(ctx context.Context, steamID string, year uint32, opts *GetFriendsSharedYearInReviewOptions) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}
	if year == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "year must be greater than zero", nil, nil)
	}
	query.Set("year", strconv.FormatUint(uint64(year), 10))
	if opts != nil && opts.ReturnPrivate != nil {
		query.Set("return_private", boolString(*opts.ReturnPrivate))
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SaleFeatureServiceGetFriendsSharedYearInReview,
		Query:  query,
	})
}

// GetUserYearAchievementsOptions controls optional query parameters for GetUserYearAchievements.
type GetUserYearAchievementsOptions struct {
	SteamID   string
	Year      uint32
	AppIDs    []uint32
	TotalOnly *bool
}

// GetUserYearAchievements returns Year in Review achievement summaries for the authenticated user.
func (s *Service) GetUserYearAchievements(ctx context.Context, accessToken string, opts *GetUserYearAchievementsOptions) (GetUserYearAchievementsResponse, error) {
	body, err := s.GetUserYearAchievementsRaw(ctx, accessToken, opts)
	if err != nil {
		return GetUserYearAchievementsResponse{}, err
	}
	return response.DecodeJSON[GetUserYearAchievementsResponse](body)
}

// GetUserYearAchievementsRaw returns the raw JSON response body.
func (s *Service) GetUserYearAchievementsRaw(ctx context.Context, accessToken string, opts *GetUserYearAchievementsOptions) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "user year achievements options are required", nil, nil)
	}
	steamID := strings.TrimSpace(opts.SteamID)
	if steamID == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id is required", nil, nil)
	}
	if opts.Year == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "year must be greater than zero", nil, nil)
	}
	query.Set("steamid", steamID)
	query.Set("year", strconv.FormatUint(uint64(opts.Year), 10))
	for idx, appID := range opts.AppIDs {
		if appID == 0 {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
		}
		query.Set("appids["+strconv.Itoa(idx)+"]", strconv.FormatUint(uint64(appID), 10))
	}
	if opts.TotalOnly != nil {
		query.Set("total_only", boolString(*opts.TotalOnly))
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SaleFeatureServiceGetUserYearAchievements,
		Query:  query,
	})
}

// GetUserYearInReview returns the public Year in Review summary for the given player and year.
func (s *Service) GetUserYearInReview(ctx context.Context, steamID string, year uint32) (GetUserYearInReviewResponse, error) {
	body, err := s.GetUserYearInReviewRaw(ctx, steamID, year)
	if err != nil {
		return GetUserYearInReviewResponse{}, err
	}
	return response.DecodeJSON[GetUserYearInReviewResponse](body)
}

// GetUserYearInReviewRaw returns the raw JSON response body.
func (s *Service) GetUserYearInReviewRaw(ctx context.Context, steamID string, year uint32) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}
	if year == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "year must be greater than zero", nil, nil)
	}
	query.Set("year", strconv.FormatUint(uint64(year), 10))

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SaleFeatureServiceGetUserYearInReview,
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

func buildAccessTokenQuery(accessToken string) (url.Values, error) {
	trimmed := strings.TrimSpace(accessToken)
	if trimmed == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "access token is required", nil, nil)
	}
	query := url.Values{}
	query.Set("access_token", trimmed)
	return query, nil
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
