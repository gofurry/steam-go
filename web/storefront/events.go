package storefront

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

const (
	defaultAdjacentPartnerEventsCountBefore = 1
	defaultAdjacentPartnerEventsCountAfter  = 10
)

// GetAdjacentPartnerEventsOptions controls optional query parameters for adjacent partner events.
type GetAdjacentPartnerEventsOptions struct {
	CountBefore  int
	CountAfter   int
	LanguageList string
}

// GetAdjacentPartnerEvents returns Storefront adjacent partner events for one AppID.
func (s *Service) GetAdjacentPartnerEvents(ctx context.Context, appID uint32, opts *GetAdjacentPartnerEventsOptions) (AdjacentPartnerEventsResponse, error) {
	body, err := s.GetAdjacentPartnerEventsRaw(ctx, appID, opts)
	if err != nil {
		return AdjacentPartnerEventsResponse{}, err
	}
	return response.DecodeJSON[AdjacentPartnerEventsResponse](body)
}

// GetAdjacentPartnerEventsRaw returns the raw JSON response body for adjacent partner events.
func (s *Service) GetAdjacentPartnerEventsRaw(ctx context.Context, appID uint32, opts *GetAdjacentPartnerEventsOptions) ([]byte, error) {
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}

	countBefore := defaultAdjacentPartnerEventsCountBefore
	countAfter := defaultAdjacentPartnerEventsCountAfter
	languageList := ""
	if opts != nil {
		if opts.CountBefore < 0 {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "count before must not be negative", nil, nil)
		}
		if opts.CountAfter < 0 {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "count after must not be negative", nil, nil)
		}
		if opts.CountBefore > 0 {
			countBefore = opts.CountBefore
		}
		if opts.CountAfter > 0 {
			countAfter = opts.CountAfter
		}
		languageList = strings.TrimSpace(opts.LanguageList)
	}

	query := url.Values{}
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))
	query.Set("count_before", strconv.Itoa(countBefore))
	query.Set("count_after", strconv.Itoa(countAfter))
	if languageList != "" {
		query.Set("lang_list", languageList)
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method:       http.MethodGet,
		Path:         webendpoint.StoreAdjacentPartnerEventsPath,
		Query:        query,
		TrafficClass: itraffic.ClassPublicStorePage,
	})
}
