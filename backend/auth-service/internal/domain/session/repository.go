package session

import (
	"context"
	"time"
)

type Repository interface {
	SaveRefreshToken(ctx context.Context, userID, refreshToken string, expiresIn time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (string, error)
	DeleteRefreshToken(ctx context.Context, refreshToken string) error
}
