package storebrowseservice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofurry/steam-go/internal/endpoint"
	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	"github.com/gofurry/steam-go/internal/response"
)

// GetContentHubConfig returns Steam content hub configuration metadata.
func (s *Service) GetContentHubConfig(ctx context.Context) (GetContentHubConfigResponse, error) {
	body, err := s.GetContentHubConfigRaw(ctx)
	if err != nil {
		return GetContentHubConfigResponse{}, err
	}
	return response.DecodeJSON[GetContentHubConfigResponse](body)
}

// GetContentHubConfigRaw returns the raw JSON response body.
func (s *Service) GetContentHubConfigRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.StoreBrowseServiceGetContentHubConfig,
	})
}

// GetItems returns Store item metadata for the supplied item IDs.
func (s *Service) GetItems(ctx context.Context, req GetItemsRequest) (GetItemsResponse, error) {
	body, err := s.GetItemsRaw(ctx, req)
	if err != nil {
		return GetItemsResponse{}, err
	}
	return response.DecodeJSON[GetItemsResponse](body)
}

// GetItemsRaw returns the raw JSON response body.
func (s *Service) GetItemsRaw(ctx context.Context, req GetItemsRequest) ([]byte, error) {
	query, err := buildGetItemsInputJSON(req)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.StoreBrowseServiceGetItems,
		Query:  query,
	})
}

func buildGetItemsInputJSON(req GetItemsRequest) (url.Values, error) {
	if len(req.IDs) == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "ids must not be empty", nil, nil)
	}
	for _, id := range req.IDs {
		if id.AppID == 0 && id.PackageID == 0 && id.BundleID == 0 {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "one of appid, packageid, or bundleid must be greater than zero", nil, nil)
		}
	}

	cloned := req
	cloned.IDs = append([]StoreItemID(nil), req.IDs...)
	if req.Context != nil {
		contextClone := *req.Context
		contextClone.CountryCode = strings.TrimSpace(contextClone.CountryCode)
		contextClone.Language = strings.TrimSpace(contextClone.Language)
		cloned.Context = &contextClone
	}
	if req.DataRequest != nil {
		dataRequestClone := *req.DataRequest
		cloned.DataRequest = &dataRequestClone
	}

	body, err := json.Marshal(cloned)
	if err != nil {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "marshal input_json failed", nil, err)
	}

	query := url.Values{}
	query.Set("input_json", string(body))
	return query, nil
}
