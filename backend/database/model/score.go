package model

import "gorm.io/gorm"

type Score struct {
	gorm.Model
	ID       uint
	UserId   uint   `gorm:"not null"`
	Score    uint64 `gorm:"not null"`
	MaxCombo uint64 `gorm:"not null"`
}
