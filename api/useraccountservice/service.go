package useraccountservice

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes IUserAccountService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a UserAccountService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
