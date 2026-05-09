package steamapps

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/GoFurry/steam-go/internal/endpoint"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/response"
)

// GetSDRConfig returns Steam Datagram Relay configuration for a game.
func (s *Service) GetSDRConfig(ctx context.Context, appID uint32) (GetSDRConfigResponse, error) {
	body, err := s.GetSDRConfigRaw(ctx, appID)
	if err != nil {
		return GetSDRConfigResponse{}, err
	}
	return response.DecodeJSON[GetSDRConfigResponse](body)
}

// GetSDRConfigRaw returns the raw JSON response body.
func (s *Service) GetSDRConfigRaw(ctx context.Context, appID uint32) ([]byte, error) {
	query, err := buildAppIDQuery(appID)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamAppsGetSDRConfig,
		Query:  query,
	})
}

// GetServersAtAddress returns known servers hosted at the given address.
func (s *Service) GetServersAtAddress(ctx context.Context, addr string) (GetServersAtAddressResponse, error) {
	body, err := s.GetServersAtAddressRaw(ctx, addr)
	if err != nil {
		return GetServersAtAddressResponse{}, err
	}
	return response.DecodeJSON[GetServersAtAddressResponse](body)
}

// GetServersAtAddressRaw returns the raw JSON response body.
func (s *Service) GetServersAtAddressRaw(ctx context.Context, addr string) ([]byte, error) {
	trimmed := strings.TrimSpace(addr)
	if trimmed == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "addr is required", nil, nil)
	}

	query := url.Values{}
	query.Set("addr", trimmed)

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamAppsGetServersAtAddress,
		Query:  query,
	})
}

// UpToDateCheck returns whether an installed game version is current.
func (s *Service) UpToDateCheck(ctx context.Context, appID, version uint32) (UpToDateCheckResponse, error) {
	body, err := s.UpToDateCheckRaw(ctx, appID, version)
	if err != nil {
		return UpToDateCheckResponse{}, err
	}
	return response.DecodeJSON[UpToDateCheckResponse](body)
}

// UpToDateCheckRaw returns the raw JSON response body.
func (s *Service) UpToDateCheckRaw(ctx context.Context, appID, version uint32) ([]byte, error) {
	query, err := buildAppIDQuery(appID)
	if err != nil {
		return nil, err
	}
	if version == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "version must be greater than zero", nil, nil)
	}
	query.Set("version", strconv.FormatUint(uint64(version), 10))

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamAppsUpToDateCheck,
		Query:  query,
	})
}

func buildAppIDQuery(appID uint32) (url.Values, error) {
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}
	query := url.Values{}
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))
	return query, nil
}
