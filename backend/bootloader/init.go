package bootloader

import (
	"context"
	"github.com/chitangUI/electronic-wooden-fish/config"
	"github.com/chitangUI/electronic-wooden-fish/database"
	"github.com/chitangUI/electronic-wooden-fish/logger"
	"github.com/chitangUI/electronic-wooden-fish/server"
	"go.uber.org/fx"
)

func InitApp(ctx context.Context) *fx.App {
	app := fx.New(
		fx.Supply(
			fx.Annotate(ctx, fx.As(new(context.Context))),
		),

		logger.Module(),
		fx.WithLogger(logger.FxLogger),

		config.Module(),

		server.Module(),
		database.Module(),
	)

	return app
}
