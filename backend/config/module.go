package config

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("config", fx.Provide(NewConfig))
}
