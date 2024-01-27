package logger

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module(
		"logger",
		fx.Provide(NewLogger),
	)
}
