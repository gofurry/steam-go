package userreviewsservice

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

// GetFriendsRecommendedApp returns account ids of friends who recommended the app.
func (s *Service) GetFriendsRecommendedApp(ctx context.Context, accessToken string, appID uint32) (GetFriendsRecommendedAppResponse, error) {
	body, err := s.GetFriendsRecommendedAppRaw(ctx, accessToken, appID)
	if err != nil {
		return GetFriendsRecommendedAppResponse{}, err
	}
	return response.DecodeJSON[GetFriendsRecommendedAppResponse](body)
}

// GetFriendsRecommendedAppRaw returns the raw JSON response body.
func (s *Service) GetFriendsRecommendedAppRaw(ctx context.Context, accessToken string, appID uint32) ([]byte, error) {
	query, err := buildAccessTokenAndAppIDQuery(accessToken, appID)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.UserReviewsServiceGetFriendsRecommendedApp,
		Query:  query,
	})
}

func buildAccessTokenAndAppIDQuery(accessToken string, appID uint32) (url.Values, error) {
	trimmed := strings.TrimSpace(accessToken)
	if trimmed == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "access token is required", nil, nil)
	}
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id is required", nil, nil)
	}

	query := url.Values{}
	query.Set("access_token", trimmed)
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))
	return query, nil
}
