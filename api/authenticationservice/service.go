package authenticationservice

import "github.com/gofurry/steam-go/internal/request"

// Service exposes IAuthenticationService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds an AuthenticationService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
