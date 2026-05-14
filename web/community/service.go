package community

import "github.com/gofurry/steam-go/internal/request"

// Service exposes read-only Steam Community web JSON endpoints.
type Service struct {
	executor *request.Executor
}

// NewService builds a Community service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
