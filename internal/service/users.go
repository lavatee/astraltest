package service

import (
	"context"

	"github.com/lavatee/astraltest/internal/model"
	"github.com/lavatee/astraltest/internal/repository"
)

type UserService struct {
	repo *repository.Repository
}

func NewUserService(repo *repository.Repository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	return s.repo.Users.GetByID(ctx, id)
}
