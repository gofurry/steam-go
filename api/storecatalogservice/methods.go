package storecatalogservice

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

// GetDevPageLinks returns developer page links for a specific app.
func (s *Service) GetDevPageLinks(ctx context.Context, appID uint32) (GetDevPageLinksResponse, error) {
	body, err := s.GetDevPageLinksRaw(ctx, appID)
	if err != nil {
		return GetDevPageLinksResponse{}, err
	}
	return response.DecodeJSON[GetDevPageLinksResponse](body)
}

// GetDevPageLinksRaw returns the raw JSON response body.
func (s *Service) GetDevPageLinksRaw(ctx context.Context, appID uint32) ([]byte, error) {
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.StoreCatalogServiceGetDevPageLinks,
		Query:  query,
	})
}
