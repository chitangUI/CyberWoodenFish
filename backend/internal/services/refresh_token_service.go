package services

import (
	"electronic-muyu-backend/internal/models"
	"time"

	"gorm.io/gorm"
)

type RefreshTokenService struct {
	db *gorm.DB
}

func NewRefreshTokenService(db *gorm.DB) *RefreshTokenService {
	return &RefreshTokenService{
		db: db,
	}
}

// Store stores a refresh token for a user
func (s *RefreshTokenService) Store(userID uint, token string, expiresAt time.Time) error {
	refreshToken := models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	return s.db.Create(&refreshToken).Error
}

// Validate validates a refresh token and returns the user ID
func (s *RefreshTokenService) Validate(token string) (uint, error) {
	var refreshToken models.RefreshToken

	err := s.db.Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&refreshToken).Error

	if err != nil {
		return 0, err
	}

	return refreshToken.UserID, nil
}

// Revoke revokes a specific refresh token
func (s *RefreshTokenService) Revoke(token string) error {
	return s.db.Where("token = ?", token).Delete(&models.RefreshToken{}).Error
}

// RevokeAllForUser revokes all refresh tokens for a specific user
func (s *RefreshTokenService) RevokeAllForUser(userID uint) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}

// CleanupExpired removes expired refresh tokens from the database
func (s *RefreshTokenService) CleanupExpired() error {
	return s.db.Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{}).Error
}
