package steamnotificationservice

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

// GetSteamNotificationsOptions controls optional query parameters for GetSteamNotifications.
type GetSteamNotificationsOptions struct {
	IncludeHidden            *bool
	Language                 *int32
	IncludeConfirmationCount *bool
	IncludePinnedCounts      *bool
	IncludeRead              *bool
	CountOnly                *bool
}

// GetPreferences returns the caller's notification preferences.
func (s *Service) GetPreferences(ctx context.Context, accessToken string) (GetPreferencesResponse, error) {
	body, err := s.GetPreferencesRaw(ctx, accessToken)
	if err != nil {
		return GetPreferencesResponse{}, err
	}
	return response.DecodeJSON[GetPreferencesResponse](body)
}

// GetPreferencesRaw returns the raw JSON response body.
func (s *Service) GetPreferencesRaw(ctx context.Context, accessToken string) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamNotificationServiceGetPreferences,
		Query:  query,
	})
}

// GetSteamNotifications returns the caller's Steam notifications.
func (s *Service) GetSteamNotifications(ctx context.Context, accessToken string, opts *GetSteamNotificationsOptions) (GetSteamNotificationsResponse, error) {
	body, err := s.GetSteamNotificationsRaw(ctx, accessToken, opts)
	if err != nil {
		return GetSteamNotificationsResponse{}, err
	}
	return response.DecodeJSON[GetSteamNotificationsResponse](body)
}

// GetSteamNotificationsRaw returns the raw JSON response body.
func (s *Service) GetSteamNotificationsRaw(ctx context.Context, accessToken string, opts *GetSteamNotificationsOptions) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		if opts.IncludeHidden != nil {
			query.Set("include_hidden", strconv.FormatBool(*opts.IncludeHidden))
		}
		if opts.Language != nil {
			query.Set("language", strconv.FormatInt(int64(*opts.Language), 10))
		}
		if opts.IncludeConfirmationCount != nil {
			query.Set("include_confirmation_count", strconv.FormatBool(*opts.IncludeConfirmationCount))
		}
		if opts.IncludePinnedCounts != nil {
			query.Set("include_pinned_counts", strconv.FormatBool(*opts.IncludePinnedCounts))
		}
		if opts.IncludeRead != nil {
			query.Set("include_read", strconv.FormatBool(*opts.IncludeRead))
		}
		if opts.CountOnly != nil {
			query.Set("count_only", strconv.FormatBool(*opts.CountOnly))
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamNotificationServiceGetSteamNotifications,
		Query:  query,
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
