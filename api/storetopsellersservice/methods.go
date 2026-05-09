package storetopsellersservice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/GoFurry/steam-go/internal/endpoint"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/response"
)

// GetCountryList returns the supported store country list.
func (s *Service) GetCountryList(ctx context.Context) (GetCountryListResponse, error) {
	body, err := s.GetCountryListRaw(ctx)
	if err != nil {
		return GetCountryListResponse{}, err
	}
	return response.DecodeJSON[GetCountryListResponse](body)
}

// GetCountryListRaw returns the raw JSON response body.
func (s *Service) GetCountryListRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.StoreTopSellersServiceGetCountryList,
	})
}

// GetWeeklyTopSellers returns the weekly top sellers payload for a country.
func (s *Service) GetWeeklyTopSellers(ctx context.Context, countryCode string) (GetWeeklyTopSellersResponse, error) {
	body, err := s.GetWeeklyTopSellersRaw(ctx, countryCode)
	if err != nil {
		return GetWeeklyTopSellersResponse{}, err
	}
	return response.DecodeJSON[GetWeeklyTopSellersResponse](body)
}

// GetWeeklyTopSellersRaw returns the raw JSON response body.
func (s *Service) GetWeeklyTopSellersRaw(ctx context.Context, countryCode string) ([]byte, error) {
	query, err := buildCountryCodeInputJSON(countryCode)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.StoreTopSellersServiceGetWeeklyTopSellers,
		Query:  query,
	})
}

func buildCountryCodeInputJSON(countryCode string) (url.Values, error) {
	trimmed := strings.TrimSpace(countryCode)
	if trimmed == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "country code is required", nil, nil)
	}

	payload := struct {
		Context struct {
			CountryCode string `json:"country_code"`
		} `json:"context"`
	}{}
	payload.Context.CountryCode = trimmed

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "marshal input_json failed", nil, err)
	}

	query := url.Values{}
	query.Set("input_json", string(body))
	return query, nil
}
