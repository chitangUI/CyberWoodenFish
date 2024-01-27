package main

import (
	"context"
	"github.com/chitangUI/electronic-wooden-fish/config"
	"github.com/chitangUI/electronic-wooden-fish/database"
	"github.com/chitangUI/electronic-wooden-fish/database/model"
	"log"
)

func main() {
	ctx := context.Background()
	gormConfig := config.NewConfig(ctx)

	db, err := database.NewDatabase(gormConfig)
	if err != nil {
		log.Fatal("can't connect your fucking database")
	}

	err = db.AutoMigrate(
		model.Score{},
		model.User{},
	)
	if err != nil {
		log.Fatal("can't run auto migrate database: ", err)
	}
}
