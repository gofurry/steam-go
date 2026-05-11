package transport

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

type requestControlManager struct {
	rateLimiter   RateLimiterConfig
	maxConcurrent int

	mu         sync.Mutex
	limiters   map[string]*rate.Limiter
	semaphores map[string]chan struct{}
}

func newRequestControlManager(cfg RequestControlConfig) *requestControlManager {
	if cfg.MaxConcurrent <= 0 && (cfg.RateLimiter.Limit <= 0 || cfg.RateLimiter.Burst <= 0) {
		return nil
	}
	return &requestControlManager{
		rateLimiter:   cfg.RateLimiter,
		maxConcurrent: cfg.MaxConcurrent,
	}
}

func acquireRequestControl(ctx context.Context, manager *requestControlManager, key string) (func(), error) {
	if manager == nil || manager.maxConcurrent <= 0 || key == "" {
		return func() {}, nil
	}

	semaphore := manager.semaphore(key)
	select {
	case semaphore <- struct{}{}:
		return func() {
			select {
			case <-semaphore:
			default:
			}
		}, nil
	case <-ctx.Done():
		return func() {}, ctx.Err()
	}
}

func waitRequestControl(ctx context.Context, manager *requestControlManager, key string) error {
	if manager == nil || key == "" {
		return nil
	}
	if manager.rateLimiter.Limit <= 0 || manager.rateLimiter.Burst <= 0 {
		return nil
	}
	return manager.limiter(key).Wait(ctx)
}

func (m *requestControlManager) limiter(key string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.limiters == nil {
		m.limiters = make(map[string]*rate.Limiter)
	}
	if limiter, ok := m.limiters[key]; ok {
		return limiter
	}

	limiter := rate.NewLimiter(m.rateLimiter.Limit, m.rateLimiter.Burst)
	m.limiters[key] = limiter
	return limiter
}

func (m *requestControlManager) semaphore(key string) chan struct{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.semaphores == nil {
		m.semaphores = make(map[string]chan struct{})
	}
	if semaphore, ok := m.semaphores[key]; ok {
		return semaphore
	}

	semaphore := make(chan struct{}, m.maxConcurrent)
	m.semaphores[key] = semaphore
	return semaphore
}
