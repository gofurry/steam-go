package steamwebapiutil

import (
	"context"
	"net/http"

	"github.com/GoFurry/steam-go/internal/endpoint"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/response"
)

// GetServerInfo returns current Steam Web API server time metadata.
func (s *Service) GetServerInfo(ctx context.Context) (GetServerInfoResponse, error) {
	body, err := s.GetServerInfoRaw(ctx)
	if err != nil {
		return GetServerInfoResponse{}, err
	}
	return response.DecodeJSON[GetServerInfoResponse](body)
}

// GetServerInfoRaw returns the raw JSON response body.
func (s *Service) GetServerInfoRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamWebAPIUtilGetServerInfo,
	})
}

// GetSupportedAPIList returns Steam's supported interface metadata.
func (s *Service) GetSupportedAPIList(ctx context.Context) (GetSupportedAPIListResponse, error) {
	body, err := s.GetSupportedAPIListRaw(ctx)
	if err != nil {
		return GetSupportedAPIListResponse{}, err
	}
	return response.DecodeJSON[GetSupportedAPIListResponse](body)
}

// GetSupportedAPIListRaw returns the raw JSON response body.
func (s *Service) GetSupportedAPIListRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamWebAPIUtilGetSupportedAPIList,
	})
}
