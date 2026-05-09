package userstorevisitservice

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

// GetMostVisitedItemsOnStoreOptions controls optional data_request fields for GetMostVisitedItemsOnStore.
type GetMostVisitedItemsOnStoreOptions struct {
	IncludeAssets                 *bool  `json:"include_assets,omitempty"`
	IncludeRelease                *bool  `json:"include_release,omitempty"`
	IncludePlatforms              *bool  `json:"include_platforms,omitempty"`
	IncludeAllPurchaseOptions     *bool  `json:"include_all_purchase_options,omitempty"`
	IncludeScreenshots            *bool  `json:"include_screenshots,omitempty"`
	IncludeTrailers               *bool  `json:"include_trailers,omitempty"`
	IncludeRatings                *bool  `json:"include_ratings,omitempty"`
	IncludeTagCount               string `json:"include_tag_count,omitempty"`
	IncludeReviews                *bool  `json:"include_reviews,omitempty"`
	IncludeBasicInfo              *bool  `json:"include_basic_info,omitempty"`
	IncludeSupportedLanguages     *bool  `json:"include_supported_languages,omitempty"`
	IncludeFullDescription        *bool  `json:"include_full_description,omitempty"`
	IncludeIncludedItems          *bool  `json:"include_included_items,omitempty"`
	IncludeAssetsWithoutOverrides *bool  `json:"include_assets_without_overrides,omitempty"`
	ApplyUserFilters              *bool  `json:"apply_user_filters,omitempty"`
	IncludeLinks                  *bool  `json:"include_links,omitempty"`
}

// GetFrequentlyVisitedPages returns the caller's recent page visits.
func (s *Service) GetFrequentlyVisitedPages(ctx context.Context, accessToken string) (GetFrequentlyVisitedPagesResponse, error) {
	body, err := s.GetFrequentlyVisitedPagesRaw(ctx, accessToken)
	if err != nil {
		return GetFrequentlyVisitedPagesResponse{}, err
	}
	return response.DecodeJSON[GetFrequentlyVisitedPagesResponse](body)
}

// GetFrequentlyVisitedPagesRaw returns the raw JSON response body.
func (s *Service) GetFrequentlyVisitedPagesRaw(ctx context.Context, accessToken string) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.UserStoreVisitServiceGetFrequentlyVisitedPages,
		Query:  query,
	})
}

// GetMostVisitedItemsOnStore returns detailed store items for the most-visited store entries in a country.
func (s *Service) GetMostVisitedItemsOnStore(ctx context.Context, countryCode string, opts *GetMostVisitedItemsOnStoreOptions) (GetMostVisitedItemsOnStoreResponse, error) {
	body, err := s.GetMostVisitedItemsOnStoreRaw(ctx, countryCode, opts)
	if err != nil {
		return GetMostVisitedItemsOnStoreResponse{}, err
	}
	return response.DecodeJSON[GetMostVisitedItemsOnStoreResponse](body)
}

// GetMostVisitedItemsOnStoreRaw returns the raw JSON response body.
func (s *Service) GetMostVisitedItemsOnStoreRaw(ctx context.Context, countryCode string, opts *GetMostVisitedItemsOnStoreOptions) ([]byte, error) {
	query, err := buildMostVisitedItemsInputJSON(countryCode, opts)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.UserStoreVisitServiceGetMostVisitedItemsOnStore,
		Query:  query,
	})
}

func buildAccessTokenQuery(accessToken string) (url.Values, error) {
	trimmed := strings.TrimSpace(accessToken)
	if trimmed == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "access token is required", nil, nil)
	}

	query := url.Values{}
	query.Set("access_token", trimmed)
	return query, nil
}

func buildMostVisitedItemsInputJSON(countryCode string, opts *GetMostVisitedItemsOnStoreOptions) (url.Values, error) {
	trimmedCountryCode := strings.TrimSpace(countryCode)
	if trimmedCountryCode == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "country code is required", nil, nil)
	}

	requestPayload := struct {
		Context struct {
			CountryCode string `json:"country_code"`
		} `json:"context"`
		DataRequest *GetMostVisitedItemsOnStoreOptions `json:"data_request,omitempty"`
	}{}
	requestPayload.Context.CountryCode = trimmedCountryCode
	if opts != nil {
		cloned := *opts
		cloned.IncludeTagCount = strings.TrimSpace(cloned.IncludeTagCount)
		requestPayload.DataRequest = &cloned
	}

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "marshal input_json failed", nil, err)
	}

	query := url.Values{}
	query.Set("input_json", string(body))
	return query, nil
}
