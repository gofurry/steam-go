package gameserversservice

import "github.com/gofurry/steam-go/internal/request"

// Service exposes IGameServersService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a GameServersService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
