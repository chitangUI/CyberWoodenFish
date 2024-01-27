package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID          uint
	Username    string `gorm:"not null;unique"`
	Password    string `gorm:"not null"`
	TelegramId  uint64 `gorm:"not null;unique"`
	ActiveStats uint8  `gorm:"not null"`
}
