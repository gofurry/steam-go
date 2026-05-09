package steamchartsservice

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/GoFurry/steam-go/internal/endpoint"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/response"
)

// GetMonthTopAppReleasesOptions controls optional query parameters for GetMonthTopAppReleases.
type GetMonthTopAppReleasesOptions struct {
	RTimeMonth      *uint32
	IncludeDLC      *bool
	TopResultsLimit *uint32
}

// GetYearTopAppReleasesOptions controls optional query parameters for GetYearTopAppReleases.
type GetYearTopAppReleasesOptions struct {
	RTimeYear       *uint32
	IncludeDLC      *bool
	TopResultsLimit *uint32
}

// GetBestOfYearPages returns all Best of Steam yearly landing pages.
func (s *Service) GetBestOfYearPages(ctx context.Context) (GetBestOfYearPagesResponse, error) {
	body, err := s.GetBestOfYearPagesRaw(ctx)
	if err != nil {
		return GetBestOfYearPagesResponse{}, err
	}
	return response.DecodeJSON[GetBestOfYearPagesResponse](body)
}

// GetBestOfYearPagesRaw returns the raw JSON response body.
func (s *Service) GetBestOfYearPagesRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamChartsServiceGetBestOfYearPages,
	})
}

// GetGamesByConcurrentPlayers returns the current top games by concurrent players.
func (s *Service) GetGamesByConcurrentPlayers(ctx context.Context) (GetGamesByConcurrentPlayersResponse, error) {
	body, err := s.GetGamesByConcurrentPlayersRaw(ctx)
	if err != nil {
		return GetGamesByConcurrentPlayersResponse{}, err
	}
	return response.DecodeJSON[GetGamesByConcurrentPlayersResponse](body)
}

// GetGamesByConcurrentPlayersRaw returns the raw JSON response body.
func (s *Service) GetGamesByConcurrentPlayersRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamChartsServiceGetGamesByConcurrentPlayers,
	})
}

// GetMonthTopAppReleases returns the top app releases for a given month window.
func (s *Service) GetMonthTopAppReleases(ctx context.Context, opts *GetMonthTopAppReleasesOptions) (GetMonthTopAppReleasesResponse, error) {
	body, err := s.GetMonthTopAppReleasesRaw(ctx, opts)
	if err != nil {
		return GetMonthTopAppReleasesResponse{}, err
	}
	return response.DecodeJSON[GetMonthTopAppReleasesResponse](body)
}

// GetMonthTopAppReleasesRaw returns the raw JSON response body.
func (s *Service) GetMonthTopAppReleasesRaw(ctx context.Context, opts *GetMonthTopAppReleasesOptions) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamChartsServiceGetMonthTopAppReleases,
		Query:  buildMonthTopAppReleasesQuery(opts),
	})
}

// GetMostPlayedGames returns the current most-played game leaderboard.
func (s *Service) GetMostPlayedGames(ctx context.Context) (GetMostPlayedGamesResponse, error) {
	body, err := s.GetMostPlayedGamesRaw(ctx)
	if err != nil {
		return GetMostPlayedGamesResponse{}, err
	}
	return response.DecodeJSON[GetMostPlayedGamesResponse](body)
}

// GetMostPlayedGamesRaw returns the raw JSON response body.
func (s *Service) GetMostPlayedGamesRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamChartsServiceGetMostPlayedGames,
	})
}

// GetTopReleasesPages returns all monthly top-release landing pages.
func (s *Service) GetTopReleasesPages(ctx context.Context) (GetTopReleasesPagesResponse, error) {
	body, err := s.GetTopReleasesPagesRaw(ctx)
	if err != nil {
		return GetTopReleasesPagesResponse{}, err
	}
	return response.DecodeJSON[GetTopReleasesPagesResponse](body)
}

// GetTopReleasesPagesRaw returns the raw JSON response body.
func (s *Service) GetTopReleasesPagesRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamChartsServiceGetTopReleasesPages,
	})
}

// GetYearTopAppReleases returns the top app releases for a given year window.
func (s *Service) GetYearTopAppReleases(ctx context.Context, opts *GetYearTopAppReleasesOptions) (GetYearTopAppReleasesResponse, error) {
	body, err := s.GetYearTopAppReleasesRaw(ctx, opts)
	if err != nil {
		return GetYearTopAppReleasesResponse{}, err
	}
	return response.DecodeJSON[GetYearTopAppReleasesResponse](body)
}

// GetYearTopAppReleasesRaw returns the raw JSON response body.
func (s *Service) GetYearTopAppReleasesRaw(ctx context.Context, opts *GetYearTopAppReleasesOptions) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamChartsServiceGetYearTopAppReleases,
		Query:  buildYearTopAppReleasesQuery(opts),
	})
}

func buildMonthTopAppReleasesQuery(opts *GetMonthTopAppReleasesOptions) url.Values {
	query := url.Values{}
	if opts == nil {
		return query
	}
	if opts.RTimeMonth != nil {
		query.Set("rtime_month", strconv.FormatUint(uint64(*opts.RTimeMonth), 10))
	}
	if opts.IncludeDLC != nil {
		query.Set("include_dlc", boolString(*opts.IncludeDLC))
	}
	if opts.TopResultsLimit != nil {
		query.Set("top_results_limit", strconv.FormatUint(uint64(*opts.TopResultsLimit), 10))
	}
	return query
}

func buildYearTopAppReleasesQuery(opts *GetYearTopAppReleasesOptions) url.Values {
	query := url.Values{}
	if opts == nil {
		return query
	}
	if opts.RTimeYear != nil {
		query.Set("rtime_year", strconv.FormatUint(uint64(*opts.RTimeYear), 10))
	}
	if opts.IncludeDLC != nil {
		query.Set("include_dlc", boolString(*opts.IncludeDLC))
	}
	if opts.TopResultsLimit != nil {
		query.Set("top_results_limit", strconv.FormatUint(uint64(*opts.TopResultsLimit), 10))
	}
	return query
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
