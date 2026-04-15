package steamuser

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

func validateSteamIDs(steamIDs []string) (string, error) {
	if len(steamIDs) == 0 {
		return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, "at least one steam id is required", nil, nil)
	}
	if len(steamIDs) > 100 {
		return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id count must not exceed 100", nil, nil)
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
