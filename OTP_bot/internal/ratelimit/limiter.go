package ratelimit

import (
	"sync"
	"time"
)

// Limiter ограничивает максимальное число действий на ключ за окно.
type Limiter interface {
	Allow(key string) bool
}

// NoopLimiter разрешает все.
type NoopLimiter struct{}

// Allow всегда возвращает true.
func (NoopLimiter) Allow(string) bool { return true }

type bucket struct {
	count     int
	resetTime time.Time
}

// MemoryLimiter применяет ограничение по фиксированному окну в памяти.
type MemoryLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	items  map[string]bucket
}

// NewMemoryLimiter создает лимитер с заданным лимитом на окно.
func NewMemoryLimiter(limit int, window time.Duration) *MemoryLimiter {
	return &MemoryLimiter{
		limit:  limit,
		window: window,
		items:  make(map[string]bucket),
	}
}

// Allow возвращает true, когда ключ не превышает лимит.
func (l *MemoryLimiter) Allow(key string) bool {
	if l == nil {
		return true
	}
	if l.limit <= 0 {
		return true
	}

	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.items[key]
	if !ok || now.After(state.resetTime) {
		state = bucket{count: 0, resetTime: now.Add(l.window)}
	}

	if state.count >= l.limit {
		l.items[key] = state
		l.cleanupLocked(now)
		return false
	}

	state.count++
	l.items[key] = state
	l.cleanupLocked(now)
	return true
}

func (l *MemoryLimiter) cleanupLocked(now time.Time) {
	if len(l.items) < 2000 {
		return
	}
	for key, state := range l.items {
		if now.After(state.resetTime) {
			delete(l.items, key)
		}
	}
}
