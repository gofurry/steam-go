package community

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	"github.com/gofurry/steam-go/internal/response"
	"github.com/gofurry/steam-go/internal/steamid"
	itraffic "github.com/gofurry/steam-go/internal/traffic"
	"github.com/gofurry/steam-go/internal/webendpoint"
)

const defaultInventoryCount = 2000

// GetInventoryOptions controls optional query parameters for inventory lookups.
type GetInventoryOptions struct {
	Language     string
	Count        int
	StartAssetID string
}

// GetInventory returns the typed inventory payload for one SteamID/app/context tuple.
func (s *Service) GetInventory(ctx context.Context, steamID string, appID uint32, contextID string, opts *GetInventoryOptions) (InventoryResponse, error) {
	body, err := s.GetInventoryRaw(ctx, steamID, appID, contextID, opts)
	if err != nil {
		return InventoryResponse{}, err
	}
	return response.DecodeJSON[InventoryResponse](body)
}

// GetInventoryRaw returns the raw JSON response body for one inventory lookup.
func (s *Service) GetInventoryRaw(ctx context.Context, steamID string, appID uint32, contextID string, opts *GetInventoryOptions) ([]byte, error) {
	normalizedSteamID, err := steamid.ValidateSteamID64(steamID)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, err.Error(), nil, err)
	}
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}
	normalizedContextID, err := steamid.ValidateNumericID("context id", contextID)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, err.Error(), nil, err)
	}

	query := url.Values{}
	count := defaultInventoryCount
	if opts != nil {
		if language := strings.TrimSpace(opts.Language); language != "" {
			query.Set("l", language)
		}
		if opts.Count != 0 {
			if opts.Count < 1 {
				return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "count must be greater than zero", nil, nil)
			}
			count = opts.Count
		}
		if startAssetID := strings.TrimSpace(opts.StartAssetID); startAssetID != "" {
			query.Set("start_assetid", startAssetID)
		}
	}
	query.Set("count", strconv.Itoa(count))

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method:       http.MethodGet,
		Path:         fmt.Sprintf("%s/%s/%d/%s", webendpoint.CommunityInventoryPath, normalizedSteamID, appID, normalizedContextID),
		Query:        query,
		TrafficClass: itraffic.ClassCommunityWeb,
	})
}
