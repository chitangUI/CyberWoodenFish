package repository

import (
	"context"
	"github.com/chitangUI/electronic-wooden-fish/database/dal"
	"github.com/chitangUI/electronic-wooden-fish/database/model"
	"go.uber.org/fx"
)

type ScoreRepository interface {
	GetByUserId(ctx context.Context, id uint) (*model.Score, error)
	Create(ctx context.Context, score *model.Score) error
	SetScore(ctx context.Context, id uint, score *model.Score) error
}

func NewScoreRepo(repo scoreRepo) ScoreRepository {
	return &repo
}

type scoreRepo struct {
	fx.In
	Query *dal.Query
}

func (r *scoreRepo) SetScore(ctx context.Context, id uint, score *model.Score) error {
	if _, err := r.Query.WithContext(ctx).Score.Where(r.Query.Score.ID.Eq(id)).Updates(score); err != nil {
		return err
	}

	return nil
}

func (r *scoreRepo) GetByUserId(ctx context.Context, id uint) (*model.Score, error) {
	return r.Query.WithContext(ctx).Score.Where(r.Query.User.ID.Eq(id)).First()
}

func (r *scoreRepo) Create(ctx context.Context, score *model.Score) error {
	return r.Query.WithContext(ctx).Score.Create(score)
}
