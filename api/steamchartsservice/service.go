package steamchartsservice

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes ISteamChartsService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a SteamChartsService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
