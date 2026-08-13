package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/domain/session"
	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/domain/user"
	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/infrastructure/jwt"
	proto "github.com/IlyushaChic/financial-platform/backend/auth-service/proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	proto.UnimplementedAuthServiceServer
	userRepo user.Repository
	session  session.Repository
	jwtMgr   *jwt.Manager
}

func NewAuthService(userRepo user.Repository, session session.Repository, jwtMgr *jwt.Manager) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		session:  session,
		jwtMgr:   jwtMgr,
	}
}

func (s *AuthService) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}
	idBytes := make([]byte, 16)
	rand.Read(idBytes)
	userID := hex.EncodeToString(idBytes)

	u := &user.User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     req.FullName,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		if err.Error() == "user already exists" {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.RegisterResponse{
		UserId:  u.ID,
		Message: "User registered successfully",
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	u, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if u == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	accessToken, err := s.jwtMgr.GenerateAccessToken(u.ID, u.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}
	refreshToken, err := s.jwtMgr.GenerateRefreshToken(u.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}
	if err := s.session.SaveRefreshToken(ctx, u.ID, refreshToken, 720*time.Hour); err != nil {
		return nil, status.Error(codes.Internal, "failed to save session")
	}
	return &proto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtMgr.AccessDuration().Seconds()),
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	claims, err := s.jwtMgr.ValidateAccessToken(req.Token)
	if err != nil {
		return &proto.ValidateTokenResponse{Valid: false}, nil
	}
	return &proto.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID,
		Email:  claims.Email,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error) {
	userID, err := s.jwtMgr.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	storedUserID, err := s.session.GetUserIDByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, status.Error(codes.Unauthenticated, "refresh token not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if storedUserID != userID {
		return nil, status.Error(codes.Unauthenticated, "refresh token mismatch")
	}
	s.session.DeleteRefreshToken(ctx, req.RefreshToken)
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	accessToken, _ := s.jwtMgr.GenerateAccessToken(u.ID, u.Email)
	newRefreshToken, _ := s.jwtMgr.GenerateRefreshToken(u.ID)
	s.session.SaveRefreshToken(ctx, u.ID, newRefreshToken, 720*time.Hour)
	return &proto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.jwtMgr.AccessDuration().Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	err := s.session.DeleteRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return &proto.LogoutResponse{Success: false}, nil
	}
	return &proto.LogoutResponse{Success: true}, nil
}
