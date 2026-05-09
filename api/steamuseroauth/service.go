package steamuseroauth

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes ISteamUserOAuth methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a SteamUserOAuth service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
