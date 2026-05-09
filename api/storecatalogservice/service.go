package storecatalogservice

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes IStoreCatalogService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a StoreCatalogService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
