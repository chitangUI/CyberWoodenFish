package auth

import (
	"context"
	"github.com/chitangUI/electronic-wooden-fish/config"
	gormModel "github.com/chitangUI/electronic-wooden-fish/database/model"
	"github.com/chitangUI/electronic-wooden-fish/database/repository"
	"github.com/chitangUI/electronic-wooden-fish/model"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/jwt"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Controller struct {
	userRepo       repository.UserRepository
	config         *config.Config
	AuthMiddleware *jwt.HertzJWTMiddleware
}

func NewAuthController(userRepository repository.UserRepository, config *config.Config, authMiddleware *jwt.HertzJWTMiddleware) *Controller {
	return &Controller{
		userRepo:       userRepository,
		config:         config,
		AuthMiddleware: authMiddleware,
	}
}

func (u *Controller) Login(_ context.Context, c *app.RequestContext) {
	var payload model.LoginRequest
	if err := c.BindAndValidate(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ReturnError(err))
		return
	}
}

func (u *Controller) Register(ctx context.Context, c *app.RequestContext) {

	logger := logrus.WithContext(ctx)

	var payload model.RegisterRequest
	if err := c.BindAndValidate(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ReturnError(err))
		return
	}

	if u.config.ReCaptcha.Enable {
		if payload.ReCaptchaResponse == "" {
			c.JSON(http.StatusBadRequest, model.ReturnError("reCaptcha enabled but no response received"))
			logger.Debug("failed get ReCaptchaResponse")
			return
		}

		_, err := ValidateCaptcha(payload.ReCaptchaResponse, u.config.ReCaptcha.SecretKey)
		if err != nil {
			logger.Debug("failed validate captcha: ", err)
			c.JSON(http.StatusBadRequest, model.ReturnError(err))
			return
		}
	}

	err := u.userRepo.CreateUser(ctx, &gormModel.User{
		Username:   payload.Username,
		Password:   payload.Password,
		TelegramId: payload.TelegramId,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ReturnError(err))
		return
	}

	c.JSON(http.StatusOK, model.ReturnSuccess("register success", nil))
}
