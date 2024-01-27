package main

import (
	"github.com/chitangUI/electronic-wooden-fish/database/model"
	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "database/dal",
		Mode:    gen.WithQueryInterface,
	})

	g.ApplyBasic(
		model.User{},
		model.Score{},
	)

	g.Execute()
}
