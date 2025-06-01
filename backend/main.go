package main

import (
	"electronic-muyu-backend/internal/app"

	"go.uber.org/fx"
)

func main() {
	fx.New(app.Module).Run()
}
