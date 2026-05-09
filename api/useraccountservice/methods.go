package useraccountservice

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

// GetUserCountry returns the caller account's country.
func (s *Service) GetUserCountry(ctx context.Context, accessToken, steamID string) (GetUserCountryResponse, error) {
	body, err := s.GetUserCountryRaw(ctx, accessToken, steamID)
	if err != nil {
		return GetUserCountryResponse{}, err
	}
	return response.DecodeJSON[GetUserCountryResponse](body)
}

// GetUserCountryRaw returns the raw JSON response body.
func (s *Service) GetUserCountryRaw(ctx context.Context, accessToken, steamID string) ([]byte, error) {
	query, err := buildAccessTokenAndSteamIDQuery(accessToken, steamID)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodPost,
		Path:   endpoint.UserAccountServiceGetUserCountry,
		Query:  query,
	})
}

func buildAccessTokenAndSteamIDQuery(accessToken, steamID string) (url.Values, error) {
	trimmedAccessToken := strings.TrimSpace(accessToken)
	if trimmedAccessToken == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "access token is required", nil, nil)
	}

	trimmedSteamID := strings.TrimSpace(steamID)
	if trimmedSteamID == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id is required", nil, nil)
	}

	query := url.Values{}
	query.Set("access_token", trimmedAccessToken)
	query.Set("steamid", trimmedSteamID)
	return query, nil
}
