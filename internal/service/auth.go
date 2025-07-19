package service

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lavatee/astraltest/internal/model"
	"github.com/lavatee/astraltest/internal/repository"
)

const (
	salt = "2ru035c3x3w25"
)

type AuthService struct {
	repo       *repository.Repository
	adminToken string
}

func NewAuthService(repo *repository.Repository, adminToken string) *AuthService {
	return &AuthService{
		repo:       repo,
		adminToken: adminToken,
	}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (string, error) {
	if req.Token != s.adminToken {
		return "", errors.New("invalid admin token")
	}
	req.Pswd = s.hashPassword(req.Pswd)
	return s.repo.Users.Create(ctx, req)
}

func (s *AuthService) hashPassword(password string) string {
	sha := sha1.New()
	sha.Write([]byte(password))
	return fmt.Sprintf("%x", sha.Sum([]byte(salt)))
}

func (s *AuthService) Authenticate(ctx context.Context, req model.AuthRequest) (string, error) {
	req.Pswd = s.hashPassword(req.Pswd)
	user, err := s.repo.Users.GetByCredentials(ctx, req.Login, req.Pswd)
	if err != nil {
		return "", errors.New("unauthorized")
	}
	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.repo.Users.CreateSession(ctx, user.ID, token, expiresAt); err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (bool, error) {
	return s.repo.Users.ValidateToken(ctx, token)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.repo.Users.DeleteSession(ctx, token)
}
