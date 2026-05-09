package wishlistservice

import "github.com/GoFurry/steam-go/internal/request"

// Service exposes IWishlistService methods.
type Service struct {
	executor *request.Executor
}

// NewService builds a WishlistService service.
func NewService(executor *request.Executor) *Service {
	return &Service{executor: executor}
}
