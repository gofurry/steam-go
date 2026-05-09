package steamuserstats

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

// GetPlayerAchievementsOptions controls optional query parameters for GetPlayerAchievements.
type GetPlayerAchievementsOptions struct {
	Language string
}

// GetGlobalAchievementPercentagesForApp returns global achievement percentages for a game.
func (s *Service) GetGlobalAchievementPercentagesForApp(ctx context.Context, gameID uint64) (GetGlobalAchievementPercentagesForAppResponse, error) {
	body, err := s.GetGlobalAchievementPercentagesForAppRaw(ctx, gameID)
	if err != nil {
		return GetGlobalAchievementPercentagesForAppResponse{}, err
	}
	return response.DecodeJSON[GetGlobalAchievementPercentagesForAppResponse](body)
}

// GetGlobalAchievementPercentagesForAppRaw returns the raw JSON response body.
func (s *Service) GetGlobalAchievementPercentagesForAppRaw(ctx context.Context, gameID uint64) ([]byte, error) {
	if gameID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "game id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("gameid", strconv.FormatUint(gameID, 10))
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserStatsGetGlobalAchievementPercentagesForApp,
		Query:  query,
	})
}

// GetPlayerAchievements returns the typed player achievements payload.
func (s *Service) GetPlayerAchievements(ctx context.Context, steamID string, appID uint32, opts *GetPlayerAchievementsOptions) (GetPlayerAchievementsResponse, error) {
	body, err := s.GetPlayerAchievementsRaw(ctx, steamID, appID, opts)
	if err != nil {
		return GetPlayerAchievementsResponse{}, err
	}

	resp, err := response.DecodeJSON[GetPlayerAchievementsResponse](body)
	if err != nil {
		return GetPlayerAchievementsResponse{}, err
	}
	if !resp.PlayerStats.Success {
		return GetPlayerAchievementsResponse{}, sdkerrors.New(
			sdkerrors.KindAPIResponse,
			0,
			"steam api returned success=false for player achievements",
			body,
			nil,
		)
	}

	return resp, nil
}

// GetPlayerAchievementsRaw returns the raw JSON response body.
func (s *Service) GetPlayerAchievementsRaw(ctx context.Context, steamID string, appID uint32, opts *GetPlayerAchievementsOptions) ([]byte, error) {
	steamID = strings.TrimSpace(steamID)
	if steamID == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id is required", nil, nil)
	}
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("steamid", steamID)
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))
	if opts != nil {
		language := strings.TrimSpace(opts.Language)
		if language != "" {
			query.Set("l", language)
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserStatsGetPlayerAchievements,
		Query:  query,
	})
}

// GetSchemaForGame returns achievement and stat schema metadata for a game.
func (s *Service) GetSchemaForGame(ctx context.Context, appID uint32) (GetSchemaForGameResponse, error) {
	body, err := s.GetSchemaForGameRaw(ctx, appID)
	if err != nil {
		return GetSchemaForGameResponse{}, err
	}
	return response.DecodeJSON[GetSchemaForGameResponse](body)
}

// GetSchemaForGameRaw returns the raw JSON response body.
func (s *Service) GetSchemaForGameRaw(ctx context.Context, appID uint32) ([]byte, error) {
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserStatsGetSchemaForGame,
		Query:  query,
	})
}

// GetNumberOfCurrentPlayers returns current player counts for a game.
func (s *Service) GetNumberOfCurrentPlayers(ctx context.Context, appID uint32) (GetNumberOfCurrentPlayersResponse, error) {
	body, err := s.GetNumberOfCurrentPlayersRaw(ctx, appID)
	if err != nil {
		return GetNumberOfCurrentPlayersResponse{}, err
	}
	return response.DecodeJSON[GetNumberOfCurrentPlayersResponse](body)
}

// GetNumberOfCurrentPlayersRaw returns the raw JSON response body.
func (s *Service) GetNumberOfCurrentPlayersRaw(ctx context.Context, appID uint32) ([]byte, error) {
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserStatsGetNumberOfCurrentPlayers,
		Query:  query,
	})
}

// GetUserStatsForGame returns raw user stats and achievement flags for a specific game.
func (s *Service) GetUserStatsForGame(ctx context.Context, steamID string, appID uint32) (GetUserStatsForGameResponse, error) {
	body, err := s.GetUserStatsForGameRaw(ctx, steamID, appID)
	if err != nil {
		return GetUserStatsForGameResponse{}, err
	}
	return response.DecodeJSON[GetUserStatsForGameResponse](body)
}

// GetUserStatsForGameRaw returns the raw JSON response body.
func (s *Service) GetUserStatsForGameRaw(ctx context.Context, steamID string, appID uint32) ([]byte, error) {
	steamID = strings.TrimSpace(steamID)
	if steamID == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id is required", nil, nil)
	}
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("steamid", steamID)
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamUserStatsGetUserStatsForGame,
		Query:  query,
	})
}
