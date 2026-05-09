package userstorevisitservice

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes IUserStoreVisitService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a UserStoreVisitService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
