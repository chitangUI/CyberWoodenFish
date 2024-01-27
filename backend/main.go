package main

import (
	"context"
	"github.com/chitangUI/electronic-wooden-fish/bootloader"
)

func main() {
	app := bootloader.InitApp(context.Background())
	app.Run()
}
