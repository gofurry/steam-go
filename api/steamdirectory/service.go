package steamdirectory

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes ISteamDirectory methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a SteamDirectory service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
