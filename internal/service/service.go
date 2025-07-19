package service

import (
	"context"
	"mime/multipart"

	"github.com/lavatee/astraltest/internal/model"
	"github.com/lavatee/astraltest/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Auth interface {
	Register(ctx context.Context, req model.RegisterRequest) (string, error)
	Authenticate(ctx context.Context, req model.AuthRequest) (string, error)
	ValidateToken(ctx context.Context, token string) (bool, error)
	Logout(ctx context.Context, token string) error
}

type Users interface {
	GetUserByID(ctx context.Context, id int) (*model.User, error)
}

type Documents interface {
	Upload(ctx context.Context, token, meta, jsonData string, file multipart.File, header *multipart.FileHeader, isFileLoaded bool) (*model.Document, error)
	GetAll(ctx context.Context, token, login, key, value string, limit int) ([]*model.Document, error)
	GetByID(ctx context.Context, token, id string) (*model.Document, []byte, error)
	Delete(ctx context.Context, token, id string) error
}

type Service struct {
	Auth
	Users
	Documents
}

func NewService(repo *repository.Repository, adminToken string, cache *redis.Client) *Service {
	return &Service{
		Auth:      NewAuthService(repo, adminToken),
		Users:     NewUserService(repo),
		Documents: NewDocumentService(repo, cache),
	}
}
