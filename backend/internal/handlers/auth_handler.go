package handlers

import (
	"electronic-muyu-backend/internal/config"
	"electronic-muyu-backend/internal/models"
	"electronic-muyu-backend/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
	userService *services.UserService
	config      *config.Config
}

type GoogleSSORequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type AppleSSORequest struct {
	IDToken      string `json:"id_token" binding:"required"`
	AuthCode     string `json:"auth_code"`
	UserIdentity string `json:"user_identity"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func NewAuthHandler(authService *services.AuthService, userService *services.UserService, config *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		config:      config,
	}
}

func (h *AuthHandler) GoogleSSO(c *gin.Context) {
	var req GoogleSSORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证 Google ID Token
	userInfo, err := h.authService.VerifyGoogleIDToken(req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google ID token"})
		return
	}

	// 查找现有用户或创建新用户
	user, err := h.userService.GetUserByGoogleID(userInfo.ID)
	if err != nil {
		// 用户不存在，检查是否有相同邮箱的用户
		existingUser, emailErr := h.userService.GetUserByEmail(userInfo.Email)
		if emailErr == nil {
			// 邮箱已存在，关联Google ID
			existingUser.GoogleID = userInfo.ID
			if existingUser.Avatar == "" {
				existingUser.Avatar = userInfo.Picture
			}
			if err := h.userService.UpdateUser(existingUser); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
				return
			}
			user = existingUser
		} else {
			// 创建新用户
			user = &models.User{
				GoogleID: userInfo.ID,
				Email:    userInfo.Email,
				Username: userInfo.Email, // 使用邮箱作为用户名
				Nickname: userInfo.Name,
				Avatar:   userInfo.Picture,
			}
			if err := h.userService.CreateUserWithSSO(user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
		}
	}

	// 生成令牌
	token, err := h.authService.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := h.authService.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// 存储刷新令牌（7天有效期）
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := h.authService.Store(user.ID, refreshToken, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

func (h *AuthHandler) AppleSSO(c *gin.Context) {
	var req AppleSSORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证 Apple ID Token
	userInfo, err := h.authService.VerifyAppleIDToken(req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Apple ID token"})
		return
	}

	// 查找现有用户或创建新用户
	user, err := h.userService.GetUserByAppleID(userInfo.Sub)
	if err != nil {
		// 用户不存在，检查是否有相同邮箱的用户
		if userInfo.Email != "" {
			existingUser, emailErr := h.userService.GetUserByEmail(userInfo.Email)
			if emailErr == nil {
				// 邮箱已存在，关联Apple ID
				existingUser.AppleID = userInfo.Sub
				if err := h.userService.UpdateUser(existingUser); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
					return
				}
				user = existingUser
			} else {
				// 创建新用户
				user = &models.User{
					AppleID:  userInfo.Sub,
					Email:    userInfo.Email,
					Username: userInfo.Email, // 使用邮箱作为用户名
					Nickname: "Apple User",   // Apple 可能不提供姓名
				}
				if err := h.userService.CreateUserWithSSO(user); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
					return
				}
			}
		} else {
			// 没有邮箱信息，创建匿名Apple用户
			user = &models.User{
				AppleID:  userInfo.Sub,
				Username: "apple_" + userInfo.Sub[:8], // 使用Apple ID的前8位作为用户名
				Nickname: "Apple User",
			}
			if err := h.userService.CreateUserWithSSO(user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
		}
	}

	// 生成令牌
	token, err := h.authService.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := h.authService.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// 存储刷新令牌（7天有效期）
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := h.authService.Store(user.ID, refreshToken, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证 refresh token 并获取用户ID
	userID, err := h.authService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 生成新的access token
	newToken, err := h.authService.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 生成新的refresh token
	newRefreshToken, err := h.authService.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// 撤销旧的refresh token
	if err := h.authService.RevokeRefreshToken(req.RefreshToken); err != nil {
		// 记录错误但不中断流程
	}

	// 存储新的refresh token（7天有效期）
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := h.authService.Store(userID, newRefreshToken, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         newToken,
		"refresh_token": newRefreshToken,
		"user":          user,
	})
}
