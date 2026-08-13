package auth_service_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/application"
	redisrepo "github.com/IlyushaChic/financial-platform/backend/auth-service/internal/infrastructure/cache/redis"
	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/infrastructure/jwt"
	postgresrepo "github.com/IlyushaChic/financial-platform/backend/auth-service/internal/infrastructure/persistence/postgres"
	proto "github.com/IlyushaChic/financial-platform/backend/auth-service/proto"
)

func TestAuthIntegration(t *testing.T) {
	ctx := context.Background()

	// ---------- 1. PostgreSQL ----------
	pgReq := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "platform",
			"POSTGRES_PASSWORD": "secret",
			"POSTGRES_DB":       "platform",
		},
		WaitingFor: wait.ForListeningPort("5432"),
	}
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pgReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")
	pgDSN := fmt.Sprintf("postgres://platform:secret@%s:%s/platform?sslmode=disable", pgHost, pgPort.Port())

	// ---------- 2. Redis ----------
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	redisHost, _ := redisContainer.Host(ctx)
	redisPort, _ := redisContainer.MappedPort(ctx, "6379")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort.Port())

	// ---------- 3. Миграции (упрощённо) ----------
	db, err := sql.Open("postgres", pgDSN)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`)
	require.NoError(t, err)

	// ---------- 4. Инициализация auth-service ----------
	userRepo := postgresrepo.NewUserRepo(db)
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	err = rdb.Ping(ctx).Err()
	require.NoError(t, err)
	sessionRepo := redisrepo.NewSessionRepo(rdb)
	jwtMgr := jwt.NewManager("testsecret", 15*time.Minute, 720*time.Hour)

	grpcServer := grpc.NewServer()
	authSvc := application.NewAuthService(userRepo, sessionRepo, jwtMgr)
	proto.RegisterAuthServiceServer(grpcServer, authSvc)

	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	grpcPort := lis.Addr().(*net.TCPAddr).Port
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", grpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	client := proto.NewAuthServiceClient(conn)

	// ---------- 5. Тесты ----------
	t.Run("Register and Login", func(t *testing.T) {
		// Register
		regResp, err := client.Register(ctx, &proto.RegisterRequest{
			Email:    "test@example.com",
			Password: "123456",
			FullName: "Test User",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, regResp.UserId)
		assert.Equal(t, "User registered successfully", regResp.Message)

		// Login
		loginResp, err := client.Login(ctx, &proto.LoginRequest{
			Email:    "test@example.com",
			Password: "123456",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, loginResp.AccessToken)
		assert.NotEmpty(t, loginResp.RefreshToken)
		assert.Greater(t, loginResp.ExpiresIn, int64(0))

		// Проверяем Redis
		keys, err := rdb.Keys(ctx, "refresh:*").Result()
		require.NoError(t, err)
		assert.NotEmpty(t, keys)

		ttl, err := rdb.TTL(ctx, "refresh:"+loginResp.RefreshToken).Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0))
	})

	t.Run("ValidateToken", func(t *testing.T) {
		loginResp, err := client.Login(ctx, &proto.LoginRequest{
			Email:    "test@example.com",
			Password: "123456",
		})
		require.NoError(t, err)

		valResp, err := client.ValidateToken(ctx, &proto.ValidateTokenRequest{
			Token: loginResp.AccessToken,
		})
		require.NoError(t, err)
		assert.True(t, valResp.Valid)
		assert.NotEmpty(t, valResp.UserId)

		valResp2, err := client.ValidateToken(ctx, &proto.ValidateTokenRequest{
			Token: "invalid",
		})
		require.NoError(t, err)
		assert.False(t, valResp2.Valid)
	})

	t.Run("RefreshToken", func(t *testing.T) {
		loginResp, err := client.Login(ctx, &proto.LoginRequest{
			Email:    "test@example.com",
			Password: "123456",
		})
		require.NoError(t, err)
		oldRefresh := loginResp.RefreshToken

		refreshResp, err := client.RefreshToken(ctx, &proto.RefreshTokenRequest{
			RefreshToken: oldRefresh,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, refreshResp.AccessToken)
		assert.NotEmpty(t, refreshResp.RefreshToken)
		assert.NotEqual(t, oldRefresh, refreshResp.RefreshToken)

		// Старый токен должен быть удалён
		_, err = rdb.Get(ctx, "refresh:"+oldRefresh).Result()
		assert.Error(t, err)
	})

	t.Run("Logout", func(t *testing.T) {
		loginResp, err := client.Login(ctx, &proto.LoginRequest{
			Email:    "test@example.com",
			Password: "123456",
		})
		require.NoError(t, err)
		refresh := loginResp.RefreshToken

		logoutResp, err := client.Logout(ctx, &proto.LogoutRequest{
			RefreshToken: refresh,
		})
		require.NoError(t, err)
		assert.True(t, logoutResp.Success)

		_, err = rdb.Get(ctx, "refresh:"+refresh).Result()
		assert.Error(t, err)
	})
}
