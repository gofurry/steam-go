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
