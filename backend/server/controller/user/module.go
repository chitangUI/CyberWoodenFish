package user

import (
	"github.com/chitangUI/electronic-wooden-fish/server/controller/user/auth"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"controller.user",
		fx.Provide(auth.NewAuthController),
		fx.Invoke(BindRoutes),
	)
}
