package accountcartservice

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofurry/steam-go/internal/endpoint"
	"github.com/gofurry/steam-go/internal/request"
	"github.com/gofurry/steam-go/internal/response"
)

// GetCartOptions controls optional query parameters for IAccountCartService/GetCart.
type GetCartOptions struct {
	UserCountry string
}

// GetCart returns the caller's shopping cart.
func (s *Service) GetCart(ctx context.Context, opts *GetCartOptions) (GetCartResponse, error) {
	body, err := s.GetCartRaw(ctx, opts)
	if err != nil {
		return GetCartResponse{}, err
	}
	return response.DecodeJSON[GetCartResponse](body)
}

// GetCartRaw returns the raw JSON response body.
func (s *Service) GetCartRaw(ctx context.Context, opts *GetCartOptions) ([]byte, error) {
	query := url.Values{}
	if opts != nil {
		userCountry := strings.TrimSpace(opts.UserCountry)
		if userCountry != "" {
			query.Set("user_country", userCountry)
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.AccountCartServiceGetCart,
		Query:  query,
	})
}

// DeleteCart removes all items from the caller's shopping cart.
func (s *Service) DeleteCart(ctx context.Context) (DeleteCartResponse, error) {
	body, err := s.DeleteCartRaw(ctx)
	if err != nil {
		return DeleteCartResponse{}, err
	}
	return response.DecodeJSON[DeleteCartResponse](body)
}

// DeleteCartRaw returns the raw JSON response body.
func (s *Service) DeleteCartRaw(ctx context.Context) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method:    http.MethodPost,
		Path:      endpoint.AccountCartServiceDeleteCart,
		Retryable: request.Retryable(false),
	})
}
