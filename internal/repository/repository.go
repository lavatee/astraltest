package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lavatee/astraltest/internal/model"
)

type Documents interface {
	Create(ctx context.Context, token string, doc *model.Document, jsonData string, fileData []byte) error
	GetAll(ctx context.Context, token, login, key, value string, limit int) ([]*model.Document, error)
	GetByID(ctx context.Context, token, id string) (*model.Document, []byte, error)
	GetFileData(ctx context.Context, token, id string) ([]byte, error)
	Delete(ctx context.Context, token, id string) error
}

type Users interface {
	Create(ctx context.Context, req model.RegisterRequest) (string, error)
	GetByCredentials(ctx context.Context, login, password string) (*model.User, error)
	CreateSession(ctx context.Context, userID int, token string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (bool, error)
	GetByID(ctx context.Context, id int) (*model.User, error)
}

type Repository struct {
	Documents
	Users
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Documents: NewDocumentsPostgres(db),
		Users:     NewUsersPostgres(db),
	}
}
