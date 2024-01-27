package game

import (
	"context"
	"github.com/chitangUI/electronic-wooden-fish/config"
	"github.com/chitangUI/electronic-wooden-fish/database/repository"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/jwt"
)

type Controller struct {
	userRepo       repository.ScoreRepository
	config         *config.Config
	AuthMiddleware *jwt.HertzJWTMiddleware
}

func NewGameController(userRepository repository.ScoreRepository, config *config.Config, authMiddleware *jwt.HertzJWTMiddleware) *Controller {
	return &Controller{
		userRepo:       userRepository,
		config:         config,
		AuthMiddleware: authMiddleware,
	}
}

func (u *Controller) GetScore(_ context.Context, c *app.RequestContext) {

}
