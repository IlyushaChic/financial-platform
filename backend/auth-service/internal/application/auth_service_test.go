package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/domain/user"
	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/infrastructure/jwt"
	proto "github.com/IlyushaChic/financial-platform/backend/auth-service/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---------- Mocks ----------

type mockUserRepo struct {
	createFunc      func(ctx context.Context, u *user.User) error
	findByEmailFunc func(ctx context.Context, email string) (*user.User, error)
	findByIDFunc    func(ctx context.Context, id string) (*user.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, u)
	}
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*user.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

type mockSessionRepo struct {
	saveRefreshTokenFunc        func(ctx context.Context, userID, token string, exp time.Duration) error
	getUserIDByRefreshTokenFunc func(ctx context.Context, token string) (string, error)
	deleteRefreshTokenFunc      func(ctx context.Context, token string) error
}

func (m *mockSessionRepo) SaveRefreshToken(ctx context.Context, userID, token string, exp time.Duration) error {
	if m.saveRefreshTokenFunc != nil {
		return m.saveRefreshTokenFunc(ctx, userID, token, exp)
	}
	return nil
}

func (m *mockSessionRepo) GetUserIDByRefreshToken(ctx context.Context, token string) (string, error) {
	if m.getUserIDByRefreshTokenFunc != nil {
		return m.getUserIDByRefreshTokenFunc(ctx, token)
	}
	return "", errors.New("not found")
}

func (m *mockSessionRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	if m.deleteRefreshTokenFunc != nil {
		return m.deleteRefreshTokenFunc(ctx, token)
	}
	return nil
}

// ---------- Helpers ----------

func setupTest() (*AuthService, *mockUserRepo, *mockSessionRepo, *jwt.Manager) {
	userRepo := &mockUserRepo{}
	sessionRepo := &mockSessionRepo{}
	jwtMgr := jwt.NewManager("testsecret", 15*time.Minute, 720*time.Hour)
	svc := NewAuthService(userRepo, sessionRepo, jwtMgr)
	return svc, userRepo, sessionRepo, jwtMgr
}

// ---------- Tests ----------

func TestRegister(t *testing.T) {
	ctx := context.Background()
	svc, repo, _, _ := setupTest()

	t.Run("success", func(t *testing.T) {
		repo.createFunc = func(ctx context.Context, u *user.User) error {
			assert.NotEmpty(t, u.ID)
			assert.Equal(t, "test@example.com", u.Email)
			assert.NotEmpty(t, u.PasswordHash)
			return nil
		}
		req := &proto.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			FullName: "Test User",
		}
		resp, err := svc.Register(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "User registered successfully", resp.Message)
		assert.NotEmpty(t, resp.UserId)
	})

	t.Run("duplicate email", func(t *testing.T) {
		repo.createFunc = func(ctx context.Context, u *user.User) error {
			return errors.New("user already exists")
		}
		req := &proto.RegisterRequest{
			Email:    "duplicate@example.com",
			Password: "password123",
		}
		resp, err := svc.Register(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})

	t.Run("invalid request (empty email)", func(t *testing.T) {
		req := &proto.RegisterRequest{
			Email:    "",
			Password: "password123",
		}
		resp, err := svc.Register(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	svc, repo, session, _ := setupTest()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	existingUser := &user.User{
		ID:           "user-123",
		Email:        "login@example.com",
		PasswordHash: string(hashed),
		FullName:     "Login User",
	}

	t.Run("success", func(t *testing.T) {
		repo.findByEmailFunc = func(ctx context.Context, email string) (*user.User, error) {
			assert.Equal(t, "login@example.com", email)
			return existingUser, nil
		}
		session.saveRefreshTokenFunc = func(ctx context.Context, userID, token string, exp time.Duration) error {
			assert.Equal(t, "user-123", userID)
			return nil
		}
		req := &proto.LoginRequest{
			Email:    "login@example.com",
			Password: "correct-password",
		}
		resp, err := svc.Login(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Greater(t, resp.ExpiresIn, int64(0))
	})

	t.Run("wrong password", func(t *testing.T) {
		repo.findByEmailFunc = func(ctx context.Context, email string) (*user.User, error) {
			return existingUser, nil
		}
		req := &proto.LoginRequest{
			Email:    "login@example.com",
			Password: "wrong-password",
		}
		resp, err := svc.Login(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("user not found", func(t *testing.T) {
		repo.findByEmailFunc = func(ctx context.Context, email string) (*user.User, error) {
			return nil, nil
		}
		req := &proto.LoginRequest{
			Email:    "notfound@example.com",
			Password: "whatever",
		}
		resp, err := svc.Login(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}

func TestValidateToken(t *testing.T) {
	ctx := context.Background()
	svc, _, _, jwtMgr := setupTest()

	t.Run("valid token", func(t *testing.T) {
		token, _ := jwtMgr.GenerateAccessToken("user-123", "test@example.com")
		req := &proto.ValidateTokenRequest{Token: token}
		resp, err := svc.ValidateToken(ctx, req)
		require.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, "user-123", resp.UserId)
		assert.Equal(t, "test@example.com", resp.Email)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := &proto.ValidateTokenRequest{Token: "invalid-token"}
		resp, err := svc.ValidateToken(ctx, req)
		require.NoError(t, err)
		assert.False(t, resp.Valid)
		assert.Empty(t, resp.UserId)
	})
}

func TestRefreshToken(t *testing.T) {
	ctx := context.Background()
	svc, repo, session, jwtMgr := setupTest()

	refreshToken, _ := jwtMgr.GenerateRefreshToken("user-123")
	userID := "user-123"

	t.Run("success", func(t *testing.T) {
		session.getUserIDByRefreshTokenFunc = func(ctx context.Context, token string) (string, error) {
			assert.Equal(t, refreshToken, token)
			return userID, nil
		}
		session.deleteRefreshTokenFunc = func(ctx context.Context, token string) error {
			assert.Equal(t, refreshToken, token)
			return nil
		}
		session.saveRefreshTokenFunc = func(ctx context.Context, userID, token string, exp time.Duration) error {
			assert.Equal(t, "user-123", userID)
			return nil
		}
		repo.findByIDFunc = func(ctx context.Context, id string) (*user.User, error) {
			assert.Equal(t, userID, id)
			return &user.User{ID: userID, Email: "test@example.com"}, nil
		}
		req := &proto.RefreshTokenRequest{RefreshToken: refreshToken}
		resp, err := svc.RefreshToken(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Greater(t, resp.ExpiresIn, int64(0))
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		req := &proto.RefreshTokenRequest{RefreshToken: "invalid"}
		resp, err := svc.RefreshToken(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})
}

func TestLogout(t *testing.T) {
	ctx := context.Background()
	svc, _, session, _ := setupTest()

	t.Run("success", func(t *testing.T) {
		session.deleteRefreshTokenFunc = func(ctx context.Context, token string) error {
			assert.Equal(t, "valid-refresh", token)
			return nil
		}
		req := &proto.LogoutRequest{RefreshToken: "valid-refresh"}
		resp, err := svc.Logout(ctx, req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("delete fails", func(t *testing.T) {
		session.deleteRefreshTokenFunc = func(ctx context.Context, token string) error {
			return errors.New("redis error")
		}
		req := &proto.LogoutRequest{RefreshToken: "any"}
		resp, err := svc.Logout(ctx, req)
		require.NoError(t, err)
		assert.False(t, resp.Success)
	})
}
