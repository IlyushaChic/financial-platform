package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepo struct {
	client *redis.Client
}

func NewSessionRepo(client *redis.Client) *SessionRepo {
	return &SessionRepo{client: client}
}

func (r *SessionRepo) SaveRefreshToken(ctx context.Context, userID, refreshToken string, expiresIn time.Duration) error {
	key := "refresh:" + refreshToken
	return r.client.Set(ctx, key, userID, expiresIn).Err()
}

func (r *SessionRepo) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	key := "refresh:" + refreshToken
	return r.client.Get(ctx, key).Result()
}

func (r *SessionRepo) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	key := "refresh:" + refreshToken
	return r.client.Del(ctx, key).Err()
}
