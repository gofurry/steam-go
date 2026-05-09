package steamdirectory

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/GoFurry/steam-go/internal/endpoint"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/response"
)

// GetCMListForConnectOptions controls optional query parameters for GetCMListForConnect.
type GetCMListForConnectOptions struct {
	CellID   *uint32
	CMType   string
	Realm    string
	MaxCount *uint32
	QOSLevel *uint32
}

// GetCMListForConnect returns connection manager candidates for a client.
func (s *Service) GetCMListForConnect(ctx context.Context, opts *GetCMListForConnectOptions) (GetCMListForConnectResponse, error) {
	body, err := s.GetCMListForConnectRaw(ctx, opts)
	if err != nil {
		return GetCMListForConnectResponse{}, err
	}
	return response.DecodeJSON[GetCMListForConnectResponse](body)
}

// GetCMListForConnectRaw returns the raw JSON response body.
func (s *Service) GetCMListForConnectRaw(ctx context.Context, opts *GetCMListForConnectOptions) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamDirectoryGetCMListForConnect,
		Query:  buildCMListForConnectQuery(opts),
	})
}

// GetSteamPipeDomains returns all known SteamPipe content domains.
func (s *Service) GetSteamPipeDomains(ctx context.Context) (GetSteamPipeDomainsResponse, error) {
	body, err := s.GetSteamPipeDomainsRaw(ctx)
	if err != nil {
		return GetSteamPipeDomainsResponse{}, err
	}
	return response.DecodeJSON[GetSteamPipeDomainsResponse](body)
}

// GetSteamPipeDomainsRaw returns the raw JSON response body.
func (s *Service) GetSteamPipeDomainsRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamDirectoryGetSteamPipeDomains,
	})
}

func buildCMListForConnectQuery(opts *GetCMListForConnectOptions) url.Values {
	query := url.Values{}
	if opts == nil {
		return query
	}
	if opts.CellID != nil {
		query.Set("cellid", strconv.FormatUint(uint64(*opts.CellID), 10))
	}
	if cmType := strings.TrimSpace(opts.CMType); cmType != "" {
		query.Set("cmtype", cmType)
	}
	if realm := strings.TrimSpace(opts.Realm); realm != "" {
		query.Set("realm", realm)
	}
	if opts.MaxCount != nil {
		query.Set("maxcount", strconv.FormatUint(uint64(*opts.MaxCount), 10))
	}
	if opts.QOSLevel != nil {
		query.Set("qoslevel", strconv.FormatUint(uint64(*opts.QOSLevel), 10))
	}
	return query
}
