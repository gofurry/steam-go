package contentserverdirectoryservice

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofurry/steam-go/internal/endpoint"
	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	"github.com/gofurry/steam-go/internal/response"
)

// GetCDNForVideo returns CDN metadata for Steam video delivery.
func (s *Service) GetCDNForVideo(ctx context.Context, req GetCDNForVideoRequest) (GetCDNForVideoResponse, error) {
	body, err := s.GetCDNForVideoRaw(ctx, req)
	if err != nil {
		return GetCDNForVideoResponse{}, err
	}
	return response.DecodeJSON[GetCDNForVideoResponse](body)
}

// GetCDNForVideoRaw returns the raw JSON response body.
func (s *Service) GetCDNForVideoRaw(ctx context.Context, req GetCDNForVideoRequest) ([]byte, error) {
	query, err := buildCDNForVideoQuery(req)
	if err != nil {
		return nil, err
	}
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.ContentServerDirectoryServiceGetCDNForVideo,
		Query:  query,
	})
}

// GetClientUpdateHosts returns Steam client update host metadata.
func (s *Service) GetClientUpdateHosts(ctx context.Context, cachedSignature string) (GetClientUpdateHostsResponse, error) {
	body, err := s.GetClientUpdateHostsRaw(ctx, cachedSignature)
	if err != nil {
		return GetClientUpdateHostsResponse{}, err
	}
	return response.DecodeJSON[GetClientUpdateHostsResponse](body)
}

// GetClientUpdateHostsRaw returns the raw JSON response body.
func (s *Service) GetClientUpdateHostsRaw(ctx context.Context, cachedSignature string) ([]byte, error) {
	query := url.Values{}
	query.Set("cached_signature", strings.TrimSpace(cachedSignature))
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.ContentServerDirectoryServiceGetClientUpdateHosts,
		Query:  query,
	})
}

// GetDepotPatchInfo returns patch metadata between two depot manifests.
func (s *Service) GetDepotPatchInfo(ctx context.Context, req GetDepotPatchInfoRequest) (GetDepotPatchInfoResponse, error) {
	body, err := s.GetDepotPatchInfoRaw(ctx, req)
	if err != nil {
		return GetDepotPatchInfoResponse{}, err
	}
	return response.DecodeJSON[GetDepotPatchInfoResponse](body)
}

// GetDepotPatchInfoRaw returns the raw JSON response body.
func (s *Service) GetDepotPatchInfoRaw(ctx context.Context, req GetDepotPatchInfoRequest) ([]byte, error) {
	query, err := buildDepotPatchInfoQuery(req)
	if err != nil {
		return nil, err
	}
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.ContentServerDirectoryServiceGetDepotPatchInfo,
		Query:  query,
	})
}

// GetServersForSteamPipe returns SteamPipe content server candidates for a cell.
func (s *Service) GetServersForSteamPipe(ctx context.Context, cellID uint32, opts *GetServersForSteamPipeOptions) (GetServersForSteamPipeResponse, error) {
	body, err := s.GetServersForSteamPipeRaw(ctx, cellID, opts)
	if err != nil {
		return GetServersForSteamPipeResponse{}, err
	}
	return response.DecodeJSON[GetServersForSteamPipeResponse](body)
}

// GetServersForSteamPipeRaw returns the raw JSON response body.
func (s *Service) GetServersForSteamPipeRaw(ctx context.Context, cellID uint32, opts *GetServersForSteamPipeOptions) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.ContentServerDirectoryServiceGetServersForSteamPipe,
		Query:  buildServersForSteamPipeQuery(cellID, opts),
	})
}

func buildCDNForVideoQuery(req GetCDNForVideoRequest) (url.Values, error) {
	clientIP := strings.TrimSpace(req.ClientIP)
	if clientIP == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "client ip must not be empty", nil, nil)
	}
	clientRegion := strings.TrimSpace(req.ClientRegion)
	if clientRegion == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "client region must not be empty", nil, nil)
	}

	query := url.Values{}
	query.Set("property_type", strconv.FormatInt(int64(req.PropertyType), 10))
	query.Set("client_ip", clientIP)
	query.Set("client_region", clientRegion)
	return query, nil
}

func buildDepotPatchInfoQuery(req GetDepotPatchInfoRequest) (url.Values, error) {
	if req.AppID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}
	if req.DepotID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "depot id must be greater than zero", nil, nil)
	}
	if req.SourceManifestID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "source manifest id must be greater than zero", nil, nil)
	}
	if req.TargetManifestID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "target manifest id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("appid", strconv.FormatUint(uint64(req.AppID), 10))
	query.Set("depotid", strconv.FormatUint(uint64(req.DepotID), 10))
	query.Set("source_manifestid", strconv.FormatUint(req.SourceManifestID, 10))
	query.Set("target_manifestid", strconv.FormatUint(req.TargetManifestID, 10))
	return query, nil
}

func buildServersForSteamPipeQuery(cellID uint32, opts *GetServersForSteamPipeOptions) url.Values {
	query := url.Values{}
	query.Set("cell_id", strconv.FormatUint(uint64(cellID), 10))
	if opts == nil {
		return query
	}
	if opts.MaxServers != nil {
		query.Set("max_servers", strconv.FormatUint(uint64(*opts.MaxServers), 10))
	}
	if ipOverride := strings.TrimSpace(opts.IPOverride); ipOverride != "" {
		query.Set("ip_override", ipOverride)
	}
	if opts.LauncherType != nil {
		query.Set("launcher_type", strconv.FormatInt(int64(*opts.LauncherType), 10))
	}
	if ipv6Public := strings.TrimSpace(opts.IPv6Public); ipv6Public != "" {
		query.Set("ipv6_public", ipv6Public)
	}
	if currentConnections := strings.TrimSpace(opts.CurrentConnections); currentConnections != "" {
		query.Set("current_connections", currentConnections)
	}
	return query
}
