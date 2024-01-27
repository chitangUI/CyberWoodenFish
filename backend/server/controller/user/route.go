package user

import (
	"fmt"
	authController "github.com/chitangUI/electronic-wooden-fish/server/controller/user/auth"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func BindRoutes(server *server.Hertz, authController *authController.Controller) {
	fmt.Println("bind server")
	api := server.Group("/api/user")

	auth := api.Group("/auth")
	auth.POST("/login", authController.AuthMiddleware.LoginHandler)
	auth.POST("/register", authController.Register)
}
