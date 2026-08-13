package clients

import (
	"context"

	authpb "github.com/IlyushaChic/financial-platform/backend/auth-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &AuthClient{
		conn:   conn,
		client: authpb.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

func (c *AuthClient) Register(ctx context.Context, email, password, fullName string) (*authpb.RegisterResponse, error) {
	req := &authpb.RegisterRequest{Email: email, Password: password, FullName: fullName}
	return c.client.Register(ctx, req)
}

func (c *AuthClient) Login(ctx context.Context, email, password string) (*authpb.LoginResponse, error) {
	req := &authpb.LoginRequest{Email: email, Password: password}
	return c.client.Login(ctx, req)
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*authpb.ValidateTokenResponse, error) {
	req := &authpb.ValidateTokenRequest{Token: token}
	return c.client.ValidateToken(ctx, req)
}

func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*authpb.RefreshTokenResponse, error) {
	req := &authpb.RefreshTokenRequest{RefreshToken: refreshToken}
	return c.client.RefreshToken(ctx, req)
}
