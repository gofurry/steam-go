package playerservice

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

// GetOwnedGamesOptions controls optional query parameters for GetOwnedGames.
type GetOwnedGamesOptions struct {
	IncludePlayedFreeGames bool
	AppIDsFilter           []uint32
}

// GetOwnedGames returns the typed owned games payload.
func (s *Service) GetOwnedGames(ctx context.Context, steamID string, opts *GetOwnedGamesOptions) (GetOwnedGamesResponse, error) {
	body, err := s.GetOwnedGamesRaw(ctx, steamID, opts)
	if err != nil {
		return GetOwnedGamesResponse{}, err
	}
	return response.DecodeJSON[GetOwnedGamesResponse](body)
}

// GetOwnedGamesRaw returns the raw JSON response body.
func (s *Service) GetOwnedGamesRaw(ctx context.Context, steamID string, opts *GetOwnedGamesOptions) ([]byte, error) {
	steamID = strings.TrimSpace(steamID)
	if steamID == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id is required", nil, nil)
	}

	query := url.Values{}
	query.Set("steamid", steamID)
	query.Set("include_appinfo", "1")
	if opts != nil {
		if opts.IncludePlayedFreeGames {
			query.Set("include_played_free_games", "1")
		}
		for _, appID := range opts.AppIDsFilter {
			query.Add("appids_filter", strconv.FormatUint(uint64(appID), 10))
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.PlayerServiceGetOwnedGames,
		Query:  query,
	})
}
