package mobilenotificationservice

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/GoFurry/steam-go/internal/endpoint"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/response"
)

// GetUserNotificationCounts returns the caller's mobile notification counters.
func (s *Service) GetUserNotificationCounts(ctx context.Context, accessToken string) (GetUserNotificationCountsResponse, error) {
	body, err := s.GetUserNotificationCountsRaw(ctx, accessToken)
	if err != nil {
		return GetUserNotificationCountsResponse{}, err
	}
	return response.DecodeJSON[GetUserNotificationCountsResponse](body)
}

// GetUserNotificationCountsRaw returns the raw JSON response body.
func (s *Service) GetUserNotificationCountsRaw(ctx context.Context, accessToken string) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.MobileNotificationServiceGetUserNotificationCounts,
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
