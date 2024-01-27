package repository

import (
	"context"
	"github.com/chitangUI/electronic-wooden-fish/database/dal"
	"github.com/chitangUI/electronic-wooden-fish/database/model"
	"go.uber.org/fx"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
}

func NewUserRepo(repo userRepo) UserRepository {
	return &repo
}

type userRepo struct {
	fx.In
	Query *dal.Query
}

func (r *userRepo) GetByID(ctx context.Context, id uint) (*model.User, error) {
	return r.Query.WithContext(ctx).User.Where(r.Query.User.ID.Eq(id)).First()
}

func (r *userRepo) CreateUser(ctx context.Context, user *model.User) error {
	return r.Query.WithContext(ctx).User.Create(user)
}
