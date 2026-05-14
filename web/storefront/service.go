package storefront

import "github.com/gofurry/steam-go/internal/request"

// Service exposes read-only Steam Storefront web JSON endpoints.
type Service struct {
	executor *request.Executor
}

// NewService builds a Storefront service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
