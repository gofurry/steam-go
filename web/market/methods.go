package market

import (
	"context"
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

const defaultCurrency = 1

// GetPriceOverviewOptions controls optional query parameters for market price overviews.
type GetPriceOverviewOptions struct {
	Currency int
}

// GetPriceOverview returns the typed market price overview payload.
func (s *Service) GetPriceOverview(ctx context.Context, appID uint32, marketHashName string, opts *GetPriceOverviewOptions) (PriceOverviewResponse, error) {
	body, err := s.GetPriceOverviewRaw(ctx, appID, marketHashName, opts)
	if err != nil {
		return PriceOverviewResponse{}, err
	}
	return response.DecodeJSON[PriceOverviewResponse](body)
}

// GetPriceOverviewRaw returns the raw JSON response body for one market price overview lookup.
func (s *Service) GetPriceOverviewRaw(ctx context.Context, appID uint32, marketHashName string, opts *GetPriceOverviewOptions) ([]byte, error) {
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}
	marketHashName = strings.TrimSpace(marketHashName)
	if marketHashName == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "market hash name must not be empty", nil, nil)
	}

	currency := defaultCurrency
	if opts != nil && opts.Currency != 0 {
		if opts.Currency < 0 {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "currency must not be negative", nil, nil)
		}
		currency = opts.Currency
	}

	query := url.Values{}
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))
	query.Set("market_hash_name", marketHashName)
	query.Set("currency", strconv.Itoa(currency))

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method:       http.MethodGet,
		Path:         webendpoint.MarketPriceOverviewPath,
		Query:        query,
		TrafficClass: itraffic.ClassMarketWeb,
	})
}
