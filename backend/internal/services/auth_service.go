package services

import (
	"crypto/rand"
	"electronic-muyu-backend/internal/config"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GoogleUserInfo represents the user information from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// AppleUserInfo represents the user information from Apple
type AppleUserInfo struct {
	Sub            string `json:"sub"`
	Email          string `json:"email"`
	EmailVerified  bool   `json:"email_verified"`
	IsPrivateEmail bool   `json:"is_private_email,omitempty"`
	RealUserStatus int    `json:"real_user_status,omitempty"`
}

type AuthService struct {
	jwtSecret        []byte
	googleClientID   string
	appleClientID    string
	refreshTokenRepo RefreshTokenRepository
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg *config.Config, refreshTokenRepo RefreshTokenRepository) *AuthService {
	return &AuthService{
		jwtSecret:        []byte(cfg.JWTSecret),
		googleClientID:   cfg.GoogleClientID,
		appleClientID:    cfg.AppleClientID,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (s *AuthService) GenerateToken(userID uint, username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ValidateTokenWithSecret validates token with a provided secret (for middleware use)
func (s *AuthService) ValidateTokenWithSecret(tokenString, jwtSecret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// VerifyGoogleIDToken verifies Google ID token and returns user info
func (s *AuthService) VerifyGoogleIDToken(idToken string) (*GoogleUserInfo, error) {
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to verify Google token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid Google ID token")
	}

	var tokenInfo struct {
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Locale        string `json:"locale"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode token info: %v", err)
	}

	// Verify the audience (client ID)
	if tokenInfo.Aud != s.googleClientID {
		return nil, errors.New("invalid audience in Google token")
	}

	return &GoogleUserInfo{
		ID:            tokenInfo.Sub,
		Email:         tokenInfo.Email,
		VerifiedEmail: tokenInfo.EmailVerified == "true",
		Name:          tokenInfo.Name,
		GivenName:     tokenInfo.GivenName,
		FamilyName:    tokenInfo.FamilyName,
		Picture:       tokenInfo.Picture,
		Locale:        tokenInfo.Locale,
	}, nil
}

// VerifyAppleIDToken verifies Apple ID token and returns user info
func (s *AuthService) VerifyAppleIDToken(idToken string) (*AppleUserInfo, error) {
	// Apple ID token verification requires parsing JWT with Apple's public keys
	// This is a simplified implementation - in production, you should:
	// 1. Fetch Apple's public keys from https://appleid.apple.com/auth/keys
	// 2. Verify the JWT signature using the appropriate key
	// 3. Validate the claims (iss, aud, exp, iat, etc.)

	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse Apple ID token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid Apple ID token claims")
	}

	// Verify audience
	if aud, ok := claims["aud"].(string); !ok || aud != s.appleClientID {
		return nil, errors.New("invalid audience in Apple token")
	}

	// Verify issuer
	if iss, ok := claims["iss"].(string); !ok || iss != "https://appleid.apple.com" {
		return nil, errors.New("invalid issuer in Apple token")
	}

	userInfo := &AppleUserInfo{}

	if sub, ok := claims["sub"].(string); ok {
		userInfo.Sub = sub
	}

	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}

	if emailVerified, ok := claims["email_verified"].(bool); ok {
		userInfo.EmailVerified = emailVerified
	}

	return userInfo, nil
}

// RefreshTokenRepository interface for refresh token storage
type RefreshTokenRepository interface {
	Store(userID uint, token string, expiresAt time.Time) error
	Validate(token string) (uint, error)
	Revoke(token string) error
	RevokeAllForUser(userID uint) error
}

// ValidateRefreshToken validates refresh token and returns user ID
func (s *AuthService) ValidateRefreshToken(token string) (uint, error) {
	return s.refreshTokenRepo.Validate(token)
}

// RevokeRefreshToken revokes a refresh token
func (s *AuthService) RevokeRefreshToken(token string) error {
	return s.refreshTokenRepo.Revoke(token)
}

// RevokeAllRefreshTokens revokes all refresh tokens for a user
func (s *AuthService) RevokeAllRefreshTokens(userID uint) error {
	return s.refreshTokenRepo.RevokeAllForUser(userID)
}

// Store stores a refresh token (delegates to repository)
func (s *AuthService) Store(userID uint, token string, expiresAt time.Time) error {
	return s.refreshTokenRepo.Store(userID, token, expiresAt)
}
