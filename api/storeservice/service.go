package storeservice

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes IStoreService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a StoreService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
