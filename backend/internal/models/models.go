package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"uniqueIndex;not null"`
	Email    string `json:"email" gorm:"uniqueIndex;not null"`
	Password string `json:"-" gorm:"not null"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`

	// SSO 相关字段
	GoogleID string `json:"-" gorm:"index"`
	AppleID  string `json:"-" gorm:"index"`

	// 游戏相关统计
	TotalScore   int64      `json:"total_score" gorm:"default:0"`
	HighestScore int64      `json:"highest_score" gorm:"default:0"`
	GamesPlayed  int        `json:"games_played" gorm:"default:0"`
	LastPlayedAt *time.Time `json:"last_played_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Score struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Score     int64     `json:"score" gorm:"not null"`
	GameMode  string    `json:"game_mode" gorm:"not null;default:'normal'"` // normal, multiplayer, challenge
	Duration  int       `json:"duration"`                                   // 游戏时长（秒）
	CreatedAt time.Time `json:"created_at"`
}

type GameSession struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	RoomID       string     `json:"room_id" gorm:"not null;index"`
	UserID       uint       `json:"user_id" gorm:"not null;index"`
	User         User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CurrentScore int64      `json:"current_score" gorm:"default:0"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	JoinedAt     time.Time  `json:"joined_at"`
	LeftAt       *time.Time `json:"left_at"`
}

type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
