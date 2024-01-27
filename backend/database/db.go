package database

import (
	"github.com/chitangUI/electronic-wooden-fish/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(config *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.GormConfig.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
