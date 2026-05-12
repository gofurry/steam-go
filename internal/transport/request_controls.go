package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type requestControlManager struct {
	rateLimiter   RateLimiterConfig
	maxConcurrent int

	mu         sync.Mutex
	limiters   map[string]*requestLimiterEntry
	semaphores map[string]*requestSemaphoreEntry

	opCount   atomic.Uint64
	lastSweep atomic.Int64
}

type requestLimiterEntry struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

type requestSemaphoreEntry struct {
	semaphore chan struct{}
	lastUsed  time.Time
}

const (
	requestControlIdleTTL          = 10 * time.Minute
	requestControlSweepIntervalOps = 128
)

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

	now := time.Now()
	manager.maybePrune(now)
	semaphore := manager.semaphore(key, now)
	select {
	case semaphore <- struct{}{}:
		return func() {
			select {
			case <-semaphore:
			default:
			}
			manager.markSemaphoreUsed(key, time.Now())
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
	now := time.Now()
	manager.maybePrune(now)
	return manager.limiter(key, now).Wait(ctx)
}

func (m *requestControlManager) limiter(key string, now time.Time) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.limiters == nil {
		m.limiters = make(map[string]*requestLimiterEntry)
	}
	if entry, ok := m.limiters[key]; ok {
		entry.lastUsed = now
		return entry.limiter
	}

	limiter := rate.NewLimiter(m.rateLimiter.Limit, m.rateLimiter.Burst)
	m.limiters[key] = &requestLimiterEntry{
		limiter:  limiter,
		lastUsed: now,
	}
	return limiter
}

func (m *requestControlManager) semaphore(key string, now time.Time) chan struct{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.semaphores == nil {
		m.semaphores = make(map[string]*requestSemaphoreEntry)
	}
	if entry, ok := m.semaphores[key]; ok {
		entry.lastUsed = now
		return entry.semaphore
	}

	semaphore := make(chan struct{}, m.maxConcurrent)
	m.semaphores[key] = &requestSemaphoreEntry{
		semaphore: semaphore,
		lastUsed:  now,
	}
	return semaphore
}

func (m *requestControlManager) markSemaphoreUsed(key string, now time.Time) {
	if m == nil || key == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.semaphores[key]; ok {
		entry.lastUsed = now
	}
}

func (m *requestControlManager) maybePrune(now time.Time) {
	if m == nil {
		return
	}
	count := m.opCount.Add(1)
	if count%requestControlSweepIntervalOps != 0 {
		return
	}
	last := m.lastSweep.Load()
	if last != 0 && now.UnixNano()-last < int64(time.Second) {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	idleBefore := now.Add(-requestControlIdleTTL)
	for key, entry := range m.limiters {
		if entry == nil || entry.lastUsed.Before(idleBefore) {
			delete(m.limiters, key)
		}
	}
	for key, entry := range m.semaphores {
		if entry == nil {
			delete(m.semaphores, key)
			continue
		}
		if len(entry.semaphore) == 0 && entry.lastUsed.Before(idleBefore) {
			delete(m.semaphores, key)
		}
	}
	m.lastSweep.Store(now.UnixNano())
}
