package steamuser

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes ISteamUser methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a SteamUser service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
