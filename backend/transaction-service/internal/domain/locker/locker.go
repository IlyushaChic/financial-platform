package locker

import (
	"context"
	"time"
)

// Locker определяет контракт для распределённой блокировки
type Locker interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (Lock, error)
}

// Lock представляет захваченную блокировку
type Lock interface {
	Unlock(ctx context.Context) error
}
