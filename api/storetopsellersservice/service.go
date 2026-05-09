package storetopsellersservice

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes IStoreTopSellersService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a StoreTopSellersService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
