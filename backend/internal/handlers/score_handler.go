package handlers

import (
	"electronic-muyu-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ScoreHandler struct {
	scoreService       *services.ScoreService
	leaderboardService *services.LeaderboardService
}

type SubmitScoreRequest struct {
	Score    int64  `json:"score" binding:"required,min=0"`
	GameMode string `json:"game_mode" binding:"required"`
	Duration int    `json:"duration" binding:"min=0"`
}

func NewScoreHandler(scoreService *services.ScoreService, leaderboardService *services.LeaderboardService) *ScoreHandler {
	return &ScoreHandler{
		scoreService:       scoreService,
		leaderboardService: leaderboardService,
	}
}

func (h *ScoreHandler) SubmitScore(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req SubmitScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score, err := h.scoreService.SubmitScore(userID.(uint), req.Score, req.GameMode, req.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit score"})
		return
	}

	c.JSON(http.StatusCreated, score)
}

func (h *ScoreHandler) GetUserScore(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	score, err := h.scoreService.GetUserBestScore(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No scores found"})
		return
	}

	totalScore, err := h.scoreService.GetUserTotalScore(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total score"})
		return
	}

	rank, err := h.leaderboardService.GetUserRank(userID.(uint))
	if err != nil {
		rank = 0 // 如果获取排名失败，设为0
	}

	c.JSON(http.StatusOK, gin.H{
		"best_score":  score,
		"total_score": totalScore,
		"rank":        rank,
	})
}

func (h *ScoreHandler) GetScoreHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	scores, err := h.scoreService.GetUserScoreHistory(userID.(uint), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get score history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scores": scores,
		"page":   page,
		"limit":  limit,
	})
}

func (h *ScoreHandler) GetLeaderboard(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	leaderboard, err := h.leaderboardService.GetGlobalLeaderboard(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
		"page":        page,
		"limit":       limit,
	})
}

func (h *ScoreHandler) GetDailyLeaderboard(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	leaderboard, err := h.leaderboardService.GetDailyLeaderboard(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get daily leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
		"page":        page,
		"limit":       limit,
		"type":        "daily",
	})
}

func (h *ScoreHandler) GetWeeklyLeaderboard(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	leaderboard, err := h.leaderboardService.GetWeeklyLeaderboard(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weekly leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
		"page":        page,
		"limit":       limit,
		"type":        "weekly",
	})
}
