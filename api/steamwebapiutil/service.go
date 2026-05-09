package steamwebapiutil

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes ISteamWebAPIUtil methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a SteamWebAPIUtil service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
