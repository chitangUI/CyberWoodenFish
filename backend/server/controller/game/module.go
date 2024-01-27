package game

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module(
		"controller.game",
		fx.Provide(NewGameController),
		fx.Invoke(BindRoutes),
	)
}
