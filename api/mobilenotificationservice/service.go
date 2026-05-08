package mobilenotificationservice

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes IMobileNotificationService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a MobileNotificationService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
