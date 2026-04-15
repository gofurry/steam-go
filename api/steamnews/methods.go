package steamnews

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GoFurry/steam-go/internal/endpoint"
	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/request"
	"github.com/GoFurry/steam-go/internal/response"
)

// GetNewsForAppOptions controls optional query parameters for ISteamNews.
type GetNewsForAppOptions struct {
	MaxLength uint32
	EndDate   time.Time
	Count     uint32
	Feeds     []string
}

// GetNewsForApp returns typed news items for the provided AppID.
func (s *Service) GetNewsForApp(ctx context.Context, appID uint32, opts *GetNewsForAppOptions) (GetNewsForAppResponse, error) {
	body, err := s.GetNewsForAppRaw(ctx, appID, opts)
	if err != nil {
		return GetNewsForAppResponse{}, err
	}
	return response.DecodeJSON[GetNewsForAppResponse](body)
}

// GetNewsForAppRaw returns the raw JSON response body.
func (s *Service) GetNewsForAppRaw(ctx context.Context, appID uint32, opts *GetNewsForAppOptions) ([]byte, error) {
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("appid", strconv.FormatUint(uint64(appID), 10))
	if opts != nil {
		if opts.MaxLength > 0 {
			query.Set("maxlength", strconv.FormatUint(uint64(opts.MaxLength), 10))
		}
		if !opts.EndDate.IsZero() {
			query.Set("enddate", strconv.FormatInt(opts.EndDate.Unix(), 10))
		}
		if opts.Count > 0 {
			query.Set("count", strconv.FormatUint(uint64(opts.Count), 10))
		}
		if len(opts.Feeds) > 0 {
			feeds := make([]string, 0, len(opts.Feeds))
			for _, feed := range opts.Feeds {
				trimmed := strings.TrimSpace(feed)
				if trimmed != "" {
					feeds = append(feeds, trimmed)
				}
			}
			if len(feeds) > 0 {
				query.Set("feeds", strings.Join(feeds, ","))
			}
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.SteamNewsGetNewsForApp,
		Query:  query,
	})
}
