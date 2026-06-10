package steamuser

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofurry/steam-go/internal/endpoint"
	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	"github.com/gofurry/steam-go/internal/response"
	"github.com/gofurry/steam-go/internal/steamid"
)

// GetFriendListOptions controls optional query parameters for GetFriendList.
type GetFriendListOptions struct {
	Relationship string
}

// GetFriendList returns the visible friend list for a Steam ID.
func (s *Service) GetFriendList(ctx context.Context, steamID string, opts *GetFriendListOptions) (GetFriendListResponse, error) {
	body, err := s.GetFriendListRaw(ctx, steamID, opts)
	if err != nil {
		return GetFriendListResponse{}, err
	}
	return response.DecodeJSON[GetFriendListResponse](body)
}

// GetFriendListRaw returns the raw JSON response body for a friend list lookup.
func (s *Service) GetFriendListRaw(ctx context.Context, steamID string, opts *GetFriendListOptions) ([]byte, error) {
	normalizedSteamID, err := validateSteamID(steamID)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("steamid", normalizedSteamID)
	if opts != nil {
		if relationship := strings.TrimSpace(opts.Relationship); relationship != "" {
			query.Set("relationship", relationship)
		}
	}
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserGetFriendList,
		Query:  query,
	})
}

// GetPlayerBans returns ban details for up to 100 Steam IDs.
func (s *Service) GetPlayerBans(ctx context.Context, steamIDs []string) (GetPlayerBansResponse, error) {
	body, err := s.GetPlayerBansRaw(ctx, steamIDs)
	if err != nil {
		return GetPlayerBansResponse{}, err
	}
	return response.DecodeJSON[GetPlayerBansResponse](body)
}

// GetPlayerBansRaw returns the raw JSON response body for player bans.
func (s *Service) GetPlayerBansRaw(ctx context.Context, steamIDs []string) ([]byte, error) {
	joined, err := validateSteamIDs(steamIDs)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("steamids", joined)
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserGetPlayerBans,
		Query:  query,
	})
}

// GetPlayerSummaries returns typed player summaries for up to 100 Steam IDs.
func (s *Service) GetPlayerSummaries(ctx context.Context, steamIDs []string) (GetPlayerSummariesResponse, error) {
	body, err := s.GetPlayerSummariesRaw(ctx, steamIDs)
	if err != nil {
		return GetPlayerSummariesResponse{}, err
	}
	return response.DecodeJSON[GetPlayerSummariesResponse](body)
}

// GetPlayerSummariesRaw returns the raw JSON response body for player summaries.
func (s *Service) GetPlayerSummariesRaw(ctx context.Context, steamIDs []string) ([]byte, error) {
	joined, err := validateSteamIDs(steamIDs)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("steamids", joined)
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserGetPlayerSummaries,
		Query:  query,
	})
}

// GetUserGroupList returns the group IDs the user belongs to.
func (s *Service) GetUserGroupList(ctx context.Context, steamID string) (GetUserGroupListResponse, error) {
	body, err := s.GetUserGroupListRaw(ctx, steamID)
	if err != nil {
		return GetUserGroupListResponse{}, err
	}
	return response.DecodeJSON[GetUserGroupListResponse](body)
}

// GetUserGroupListRaw returns the raw JSON response body for user group list lookups.
func (s *Service) GetUserGroupListRaw(ctx context.Context, steamID string) ([]byte, error) {
	normalizedSteamID, err := validateSteamID(steamID)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("steamid", normalizedSteamID)
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserGetUserGroupList,
		Query:  query,
	})
}

func validateSteamID(steamID string) (string, error) {
	normalized, err := steamid.ValidateSteamID64(steamID)
	if err != nil {
		return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, err.Error(), nil, err)
	}
	return normalized, nil
}

func validateSteamIDs(steamIDs []string) (string, error) {
	if len(steamIDs) == 0 {
		return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, "at least one steam id is required", nil, nil)
	}
	if len(steamIDs) > 100 {
		return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id count must not exceed 100", nil, nil)
	}

	normalized := make([]string, 0, len(steamIDs))
	for _, steamID := range steamIDs {
		trimmed, err := validateSteamID(steamID)
		if err != nil {
			return "", err
		}
		normalized = append(normalized, trimmed)
	}
	return strings.Join(normalized, ","), nil
}
