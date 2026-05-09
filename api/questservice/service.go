package questservice

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes IQuestService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds an IQuestService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
