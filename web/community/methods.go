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
	steamID = strings.TrimSpace(steamID)
	if steamID == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id must not be empty", nil, nil)
	}
	if err := validateNumericIdentifier("steam id", steamID); err != nil {
		return nil, err
	}
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}
	contextID = strings.TrimSpace(contextID)
	if contextID == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "context id must not be empty", nil, nil)
	}
	if err := validateNumericIdentifier("context id", contextID); err != nil {
		return nil, err
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
		Path:         fmt.Sprintf("%s/%s/%d/%s", webendpoint.CommunityInventoryPath, steamID, appID, contextID),
		Query:        query,
		TrafficClass: itraffic.ClassCommunityWeb,
	})
}

func validateNumericIdentifier(name, value string) error {
	for _, r := range value {
		if r < '0' || r > '9' {
			return sdkerrors.New(sdkerrors.KindRequestBuild, 0, fmt.Sprintf("%s must contain digits only", name), nil, nil)
		}
	}
	return nil
}
