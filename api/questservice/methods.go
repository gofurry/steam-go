package questservice

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/GoFurry/steam-go/internal/endpoint"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/response"
)

// GetCommunityInventoryOptions controls optional query parameters for GetCommunityInventory.
type GetCommunityInventoryOptions struct {
	FilterAppIDs []uint32
}

// GetCommunityInventory returns community inventory items.
func (s *Service) GetCommunityInventory(ctx context.Context, opts *GetCommunityInventoryOptions) (GetCommunityInventoryResponse, error) {
	body, err := s.GetCommunityInventoryRaw(ctx, opts)
	if err != nil {
		return GetCommunityInventoryResponse{}, err
	}
	return response.DecodeJSON[GetCommunityInventoryResponse](body)
}

// GetCommunityInventoryRaw returns the raw JSON response body.
func (s *Service) GetCommunityInventoryRaw(ctx context.Context, opts *GetCommunityInventoryOptions) ([]byte, error) {
	query := url.Values{}
	if opts != nil {
		for idx, appID := range opts.FilterAppIDs {
			if appID == 0 {
				return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
			}
			query.Set("filter_appids["+strconv.Itoa(idx)+"]", strconv.FormatUint(uint64(appID), 10))
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.QuestServiceGetCommunityInventory,
		Query:  query,
	})
}

// GetNumTradingCardsEarnedOptions controls optional query parameters for GetNumTradingCardsEarned.
type GetNumTradingCardsEarnedOptions struct {
	TimestampStart uint32
	TimestampEnd   uint32
}

// GetNumTradingCardsEarned returns the number of trading cards earned by the authenticated caller.
func (s *Service) GetNumTradingCardsEarned(ctx context.Context, accessToken string, opts *GetNumTradingCardsEarnedOptions) (GetNumTradingCardsEarnedResponse, error) {
	body, err := s.GetNumTradingCardsEarnedRaw(ctx, accessToken, opts)
	if err != nil {
		return GetNumTradingCardsEarnedResponse{}, err
	}
	return response.DecodeJSON[GetNumTradingCardsEarnedResponse](body)
}

// GetNumTradingCardsEarnedRaw returns the raw JSON response body.
func (s *Service) GetNumTradingCardsEarnedRaw(ctx context.Context, accessToken string, opts *GetNumTradingCardsEarnedOptions) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		if opts.TimestampStart > 0 {
			query.Set("timestamp_start", strconv.FormatUint(uint64(opts.TimestampStart), 10))
		}
		if opts.TimestampEnd > 0 {
			query.Set("timestamp_end", strconv.FormatUint(uint64(opts.TimestampEnd), 10))
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.QuestServiceGetNumTradingCardsEarned,
		Query:  query,
	})
}

func buildAccessTokenQuery(accessToken string) (url.Values, error) {
	if accessToken == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "access token is required", nil, nil)
	}
	query := url.Values{}
	query.Set("access_token", accessToken)
	return query, nil
}
