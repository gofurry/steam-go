package steamnews

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes ISteamNews methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a SteamNews service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
