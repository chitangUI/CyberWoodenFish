package server

import (
	"context"
	"github.com/chitangUI/electronic-wooden-fish/config"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/jwt"
	"go.uber.org/fx"
	"log"
	"time"
)

func StartServer(svr *server.Hertz, lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := svr.Run(); err != nil {
					panic(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return svr.Shutdown(ctx)
		},
	})
}

func NewServer(config *config.Config) (*server.Hertz, *jwt.HertzJWTMiddleware) {
	s := server.Default(
		server.WithHostPorts(config.HttpPort),
		server.WithKeepAlive(true),
	)

	authMiddleware, err := jwt.New(&jwt.HertzJWTMiddleware{
		Realm:      config.JwtConfig.Realm,
		Key:        []byte(config.JwtConfig.Key),
		Timeout:    4 * time.Hour,
		MaxRefresh: 4 * time.Hour,
	})
	if err != nil {
		panic(err)
	}

	if err = authMiddleware.MiddlewareInit(); err != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:", err)
	}

	return s, authMiddleware
}
