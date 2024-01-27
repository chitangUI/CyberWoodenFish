package server

import (
	"github.com/chitangUI/electronic-wooden-fish/server/controller/game"
	"github.com/chitangUI/electronic-wooden-fish/server/controller/user"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewServer),
		user.Module(),
		game.Module(),
		fx.Invoke(StartServer),
	)
}
