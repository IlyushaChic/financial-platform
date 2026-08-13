package redis

import (
	"context"
	"errors"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/locker"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// DistributedLocker реализует распределённую блокировку через Redis
type DistributedLocker struct {
	client *redis.Client
}

// NewDistributedLocker создаёт новый экземпляр
func NewDistributedLocker(client *redis.Client) *DistributedLocker {
	return &DistributedLocker{client: client}
}

// Lock пытается захватить блокировку с TTL.
// Возвращает Lock, который нужно отпустить через Unlock.
func (l *DistributedLocker) Lock(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error) {
	lockKey := "lock:" + key
	owner := uuid.New().String() // уникальный идентификатор владельца

	// SET key owner NX EX seconds – атомарно устанавливаем, только если ключа нет
	res, err := l.client.SetNX(ctx, lockKey, owner, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !res {
		return nil, errors.New("lock already held")
	}
	return &Lock{
		key:    lockKey,
		owner:  owner,
		client: l.client,
	}, nil
}

// Lock представляет захваченную блокировку
type Lock struct {
	key    string
	owner  string
	client *redis.Client
}

// Unlock освобождает блокировку, только если владелец совпадает (защита от удаления чужой блокировки)
func (l *Lock) Unlock(ctx context.Context) error {
	// Используем Lua-скрипт для атомарной проверки и удаления
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.owner).Int64()
	if err != nil {
		return err
	}
	if result == 0 {
		return errors.New("lock not owned or already released")
	}
	return nil
}
