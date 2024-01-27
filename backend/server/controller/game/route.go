package game

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func BindRoutes(server *server.Hertz, gameController *Controller) {
	fmt.Println("bind server")
	api := server.Group("/api/game")
	api.GET("/score", gameController.GetScore)
}
