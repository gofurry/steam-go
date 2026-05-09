package steamuseroauth

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

// GetUserSummaries returns typed player summaries from ISteamUserOAuth.
func (s *Service) GetUserSummaries(ctx context.Context, steamIDs []string) (GetUserSummariesResponse, error) {
	body, err := s.GetUserSummariesRaw(ctx, steamIDs)
	if err != nil {
		return GetUserSummariesResponse{}, err
	}
	return response.DecodeJSON[GetUserSummariesResponse](body)
}

// GetUserSummariesRaw returns the raw JSON response body.
func (s *Service) GetUserSummariesRaw(ctx context.Context, steamIDs []string) ([]byte, error) {
	joined, err := validateSteamIDs(steamIDs)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("steamids", joined)
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserOAuthGetUserSummaries,
		Query:  query,
	})
}

// GetFriendList returns the caller's friend list from ISteamUserOAuth.
func (s *Service) GetFriendList(ctx context.Context, accessToken string) (GetFriendListResponse, error) {
	body, err := s.GetFriendListRaw(ctx, accessToken)
	if err != nil {
		return GetFriendListResponse{}, err
	}
	return response.DecodeJSON[GetFriendListResponse](body)
}

// GetFriendListRaw returns the raw JSON response body.
func (s *Service) GetFriendListRaw(ctx context.Context, accessToken string) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserOAuthGetFriendList,
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

func validateSteamIDs(steamIDs []string) (string, error) {
	if len(steamIDs) == 0 {
		return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, "at least one steam id is required", nil, nil)
	}
	if len(steamIDs) > 100 {
		return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam ids cannot exceed 100 entries", nil, nil)
	}

	normalized := make([]string, 0, len(steamIDs))
	for _, steamID := range steamIDs {
		trimmed := strings.TrimSpace(steamID)
		if trimmed == "" {
			return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id must not be empty", nil, nil)
		}
		normalized = append(normalized, trimmed)
	}

	return strings.Join(normalized, ","), nil
}
