package database

import (
	"github.com/chitangUI/electronic-wooden-fish/database/dal"
	"github.com/chitangUI/electronic-wooden-fish/database/repository"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"database",

		fx.Provide(NewDatabase),
		fx.Provide(dal.Use),
		fx.Provide(repository.NewUserRepo),
		fx.Provide(repository.NewScoreRepo),
	)
}
