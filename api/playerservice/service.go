package playerservice

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes IPlayerService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a PlayerService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
