package storepreferencesservice

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

// GetIgnoreList returns the caller's ignored apps.
func (s *Service) GetIgnoreList(ctx context.Context, accessToken string) (GetIgnoreListResponse, error) {
	body, err := s.GetIgnoreListRaw(ctx, accessToken)
	if err != nil {
		return GetIgnoreListResponse{}, err
	}
	return response.DecodeJSON[GetIgnoreListResponse](body)
}

// GetIgnoreListRaw returns the raw JSON response body.
func (s *Service) GetIgnoreListRaw(ctx context.Context, accessToken string) ([]byte, error) {
	query, err := buildAccessTokenQuery(accessToken)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.StorePreferencesServiceGetIgnoreList,
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
