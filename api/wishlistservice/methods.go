package wishlistservice

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

// GetWishlistItemsOnSaleOptions controls optional data_request fields for GetWishlistItemsOnSale.
type GetWishlistItemsOnSaleOptions struct {
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

// GetWishlist returns the wishlist items for the provided steam id.
func (s *Service) GetWishlist(ctx context.Context, steamID string) (GetWishlistResponse, error) {
	body, err := s.GetWishlistRaw(ctx, steamID)
	if err != nil {
		return GetWishlistResponse{}, err
	}
	return response.DecodeJSON[GetWishlistResponse](body)
}

// GetWishlistRaw returns the raw JSON response body.
func (s *Service) GetWishlistRaw(ctx context.Context, steamID string) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.WishlistServiceGetWishlist,
		Query:  query,
	})
}

// GetWishlistItemCount returns the wishlist item count for the provided steam id.
func (s *Service) GetWishlistItemCount(ctx context.Context, steamID string) (GetWishlistItemCountResponse, error) {
	body, err := s.GetWishlistItemCountRaw(ctx, steamID)
	if err != nil {
		return GetWishlistItemCountResponse{}, err
	}
	return response.DecodeJSON[GetWishlistItemCountResponse](body)
}

// GetWishlistItemCountRaw returns the raw JSON response body.
func (s *Service) GetWishlistItemCountRaw(ctx context.Context, steamID string) ([]byte, error) {
	query, err := buildSteamIDQuery(steamID)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.WishlistServiceGetWishlistItemCount,
		Query:  query,
	})
}

// GetWishlistItemsOnSale returns wishlist items currently on sale for a country.
func (s *Service) GetWishlistItemsOnSale(ctx context.Context, accessToken, countryCode string, opts *GetWishlistItemsOnSaleOptions) (GetWishlistItemsOnSaleResponse, error) {
	body, err := s.GetWishlistItemsOnSaleRaw(ctx, accessToken, countryCode, opts)
	if err != nil {
		return GetWishlistItemsOnSaleResponse{}, err
	}
	return response.DecodeJSON[GetWishlistItemsOnSaleResponse](body)
}

// GetWishlistItemsOnSaleRaw returns the raw JSON response body.
func (s *Service) GetWishlistItemsOnSaleRaw(ctx context.Context, accessToken, countryCode string, opts *GetWishlistItemsOnSaleOptions) ([]byte, error) {
	query, err := buildWishlistItemsOnSaleQuery(accessToken, countryCode, opts)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.WishlistServiceGetWishlistItemsOnSale,
		Query:  query,
	})
}

func buildSteamIDQuery(steamID string) (url.Values, error) {
	trimmedSteamID := strings.TrimSpace(steamID)
	if trimmedSteamID == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id is required", nil, nil)
	}

	query := url.Values{}
	query.Set("steamid", trimmedSteamID)
	return query, nil
}

func buildWishlistItemsOnSaleQuery(accessToken, countryCode string, opts *GetWishlistItemsOnSaleOptions) (url.Values, error) {
	trimmedAccessToken := strings.TrimSpace(accessToken)
	if trimmedAccessToken == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "access token is required", nil, nil)
	}

	trimmedCountryCode := strings.TrimSpace(countryCode)
	if trimmedCountryCode == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "country code is required", nil, nil)
	}

	requestPayload := struct {
		Context struct {
			CountryCode string `json:"country_code"`
		} `json:"context"`
		DataRequest *GetWishlistItemsOnSaleOptions `json:"data_request,omitempty"`
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
	query.Set("access_token", trimmedAccessToken)
	query.Set("input_json", string(body))
	return query, nil
}
