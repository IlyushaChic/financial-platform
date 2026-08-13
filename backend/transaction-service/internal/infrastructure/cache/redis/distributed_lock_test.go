package redis

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDistributedLock(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	host, _ := redisContainer.Host(ctx)
	port, _ := redisContainer.MappedPort(ctx, "6379")
	addr := host + ":" + port.Port()

	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	locker := NewDistributedLocker(client)

	t.Run("basic lock/unlock", func(t *testing.T) {
		lock, err := locker.Lock(ctx, "test-key", 5*time.Second)
		require.NoError(t, err)
		// Приводим к конкретному типу, чтобы получить доступ к полям
		redisLock, ok := lock.(*Lock)
		require.True(t, ok, "expected *Lock")
		assert.NotEmpty(t, redisLock.owner)

		val, err := client.Get(ctx, redisLock.key).Result()
		require.NoError(t, err)
		assert.Equal(t, redisLock.owner, val)

		err = redisLock.Unlock(ctx)
		require.NoError(t, err)

		_, err = client.Get(ctx, redisLock.key).Result()
		assert.Error(t, err)
	})

	t.Run("lock already held", func(t *testing.T) {
		lock1, err := locker.Lock(ctx, "test-key2", 5*time.Second)
		require.NoError(t, err)
		defer lock1.Unlock(ctx)

		_, err = locker.Lock(ctx, "test-key2", 5*time.Second)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock already held")
	})

	t.Run("concurrent access with lock", func(t *testing.T) {
		const key = "counter"
		var wg sync.WaitGroup
		var mu sync.Mutex
		counter := 0

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var lock *Lock
				var err error
				for attempt := 0; attempt < 5; attempt++ {
					lockInterface, err := locker.Lock(ctx, key, 2*time.Second)
					if err == nil {
						// Приводим к конкретному типу
						var ok bool
						lock, ok = lockInterface.(*Lock)
						if ok {
							break
						}
					}
					time.Sleep(100 * time.Millisecond)
				}
				if err != nil || lock == nil {
					t.Errorf("failed to acquire lock after retries: %v", err)
					return
				}
				defer lock.Unlock(ctx)

				mu.Lock()
				counter++
				mu.Unlock()
				time.Sleep(50 * time.Millisecond)
			}()
		}
		wg.Wait()
		assert.Equal(t, 5, counter)
	})

	t.Run("lock expiration (TTL)", func(t *testing.T) {
		lockInterface, err := locker.Lock(ctx, "ttl-key", 1*time.Second)
		require.NoError(t, err)
		lock1, ok := lockInterface.(*Lock)
		require.True(t, ok)

		time.Sleep(1500 * time.Millisecond)

		lockInterface2, err := locker.Lock(ctx, "ttl-key", 5*time.Second)
		require.NoError(t, err)
		lock2, ok := lockInterface2.(*Lock)
		require.True(t, ok)
		defer lock2.Unlock(ctx)

		assert.NotEqual(t, lock1.owner, lock2.owner)
	})

	t.Run("unlock foreign lock fails", func(t *testing.T) {
		foreignKey := "lock:foreign"
		_, err := client.Set(ctx, foreignKey, "foreign-owner", 10*time.Second).Result()
		require.NoError(t, err)
		defer client.Del(ctx, foreignKey)

		lock := &Lock{key: foreignKey, owner: "my-owner", client: client}
		err = lock.Unlock(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock not owned")
	})
}
