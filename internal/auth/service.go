package auth

import (
	"context"
	"errors"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already in use")
)

type WorkspaceCreator interface {
	Create(ctx context.Context, name, ownerID string) (interface{}, error)
}

type Service struct {
	repo      *Repository
	wsCreator WorkspaceCreator
	jwtSecret string
}

func NewService(repo *Repository, wsCreator WorkspaceCreator) *Service {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}
	return &Service{repo: repo, wsCreator: wsCreator, jwtSecret: secret}
}

func (s *Service) Signup(ctx context.Context, req SignupRequest) (*AuthResponse, string, error) {
	user, err := s.repo.CreateUser(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		return nil, "", ErrEmailTaken
	}
	_, err = s.wsCreator.Create(ctx, user.Name+"'s Workspace", user.ID)
	if err != nil {
		return nil, "", err
	}
	accessToken, err := GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", err
	}
	refreshToken, err := s.repo.CreateRefreshToken(ctx, user.ID, time.Now().Add(7*24*time.Hour).Format(time.RFC3339))
	if err != nil {
		return nil, "", err
	}
	return &AuthResponse{User: *user, AccessToken: accessToken}, refreshToken, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, "", ErrInvalidCredentials
	}
	accessToken, err := GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", err
	}
	refreshToken, err := s.repo.CreateRefreshToken(ctx, user.ID, time.Now().Add(7*24*time.Hour).Format(time.RFC3339))
	if err != nil {
		return nil, "", err
	}
	return &AuthResponse{User: User{ID: user.ID, Email: user.Email, Name: user.Name, CreatedAt: user.CreatedAt}, AccessToken: accessToken}, refreshToken, nil
}

func (s *Service) Refresh(ctx context.Context, refreshTokenHex string) (*AuthResponse, string, error) {
	user, err := s.repo.GetUserByRefreshToken(ctx, refreshTokenHex)
	if err != nil {
		return nil, "", err
	}
	accessToken, err := GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", err
	}
	newRefresh, err := s.repo.CreateRefreshToken(ctx, user.ID, time.Now().Add(7*24*time.Hour).Format(time.RFC3339))
	if err != nil {
		return nil, "", err
	}
	return &AuthResponse{User: *user, AccessToken: accessToken}, newRefresh, nil
}

func (s *Service) Logout(ctx context.Context, userID string) error {
	return s.repo.DeleteUserRefreshTokens(ctx, userID)
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	claims, err := ValidateAccessToken(tokenString, s.jwtSecret)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func (s *Service) GetUser(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}
