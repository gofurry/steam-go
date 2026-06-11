package contentserverdirectoryservice

import "github.com/gofurry/steam-go/internal/request"

// Service exposes IContentServerDirectoryService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a ContentServerDirectoryService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
