package market

import "github.com/gofurry/steam-go/internal/request"

// Service exposes read-only Steam Market web JSON endpoints.
type Service struct {
	executor *request.Executor
}

// NewService builds a Market service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
