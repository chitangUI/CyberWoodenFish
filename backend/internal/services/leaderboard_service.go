package services

import (
	"electronic-muyu-backend/internal/models"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type LeaderboardService struct {
	db    *gorm.DB
	redis *RedisService
}

type LeaderboardEntry struct {
	UserID       uint       `json:"user_id"`
	Username     string     `json:"username"`
	Nickname     string     `json:"nickname"`
	Avatar       string     `json:"avatar"`
	Score        int64      `json:"score"`
	Rank         int        `json:"rank"`
	GamesPlayed  int        `json:"games_played"`
	LastPlayedAt *time.Time `json:"last_played_at"`
}

func NewLeaderboardService(db *gorm.DB, redis *RedisService) *LeaderboardService {
	return &LeaderboardService{
		db:    db,
		redis: redis,
	}
}

func (s *LeaderboardService) GetGlobalLeaderboard(limit int, offset int) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry

	err := s.db.Model(&models.User{}).
		Select(`
			users.id as user_id,
			users.username,
			users.nickname,
			users.avatar,
			users.highest_score as score,
			users.games_played,
			users.last_played_at,
			ROW_NUMBER() OVER (ORDER BY users.highest_score DESC) as rank
		`).
		Where("users.highest_score > 0").
		Order("users.highest_score DESC").
		Limit(limit).
		Offset(offset).
		Scan(&entries).Error

	return entries, err
}

func (s *LeaderboardService) GetDailyLeaderboard(limit int, offset int) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry
	today := time.Now().Truncate(24 * time.Hour)

	err := s.db.Table("scores").
		Select(`
			users.id as user_id,
			users.username,
			users.nickname,
			users.avatar,
			MAX(scores.score) as score,
			COUNT(scores.id) as games_played,
			MAX(scores.created_at) as last_played_at,
			ROW_NUMBER() OVER (ORDER BY MAX(scores.score) DESC) as rank
		`).
		Joins("JOIN users ON users.id = scores.user_id").
		Where("scores.created_at >= ?", today).
		Group("users.id, users.username, users.nickname, users.avatar").
		Order("score DESC").
		Limit(limit).
		Offset(offset).
		Scan(&entries).Error

	return entries, err
}

func (s *LeaderboardService) GetWeeklyLeaderboard(limit int, offset int) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry
	weekAgo := time.Now().AddDate(0, 0, -7)

	err := s.db.Table("scores").
		Select(`
			users.id as user_id,
			users.username,
			users.nickname,
			users.avatar,
			MAX(scores.score) as score,
			COUNT(scores.id) as games_played,
			MAX(scores.created_at) as last_played_at,
			ROW_NUMBER() OVER (ORDER BY MAX(scores.score) DESC) as rank
		`).
		Joins("JOIN users ON users.id = scores.user_id").
		Where("scores.created_at >= ?", weekAgo).
		Group("users.id, users.username, users.nickname, users.avatar").
		Order("score DESC").
		Limit(limit).
		Offset(offset).
		Scan(&entries).Error

	return entries, err
}

func (s *LeaderboardService) GetUserRank(userID uint) (int, error) {
	var rank int64

	err := s.db.Model(&models.User{}).
		Select("COUNT(*) + 1").
		Where("highest_score > (SELECT highest_score FROM users WHERE id = ?)", userID).
		Scan(&rank).Error

	return int(rank), err
}

// UpdateLeaderboardCache updates Redis cache with new score
func (s *LeaderboardService) UpdateLeaderboardCache(userID uint, username string, score int64) error {
	if s.redis == nil {
		return nil // Redis not available, skip caching
	}

	userData := map[string]interface{}{
		"user_id":  userID,
		"username": username,
		"score":    score,
	}

	// Add to global leaderboard sorted set
	if err := s.redis.ZAdd("leaderboard:global", float64(score), userData); err != nil {
		return fmt.Errorf("failed to update global leaderboard cache: %v", err)
	}

	// Add to daily leaderboard
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("leaderboard:daily:%s", today)
	if err := s.redis.ZAdd(dailyKey, float64(score), userData); err != nil {
		return fmt.Errorf("failed to update daily leaderboard cache: %v", err)
	}

	// Set expiration for daily leaderboard (2 days)
	s.redis.client.Expire(s.redis.ctx, dailyKey, 48*time.Hour)

	// Add to weekly leaderboard
	year, week := time.Now().ISOWeek()
	weeklyKey := fmt.Sprintf("leaderboard:weekly:%d-W%02d", year, week)
	if err := s.redis.ZAdd(weeklyKey, float64(score), userData); err != nil {
		return fmt.Errorf("failed to update weekly leaderboard cache: %v", err)
	}

	// Set expiration for weekly leaderboard (2 weeks)
	s.redis.client.Expire(s.redis.ctx, weeklyKey, 14*24*time.Hour)

	return nil
}

// GetCachedLeaderboard gets leaderboard from Redis cache
func (s *LeaderboardService) GetCachedLeaderboard(leaderboardType string, limit int, offset int) ([]LeaderboardEntry, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("Redis not available")
	}

	var key string
	switch leaderboardType {
	case "global":
		key = "leaderboard:global"
	case "daily":
		today := time.Now().Format("2006-01-02")
		key = fmt.Sprintf("leaderboard:daily:%s", today)
	case "weekly":
		year, week := time.Now().ISOWeek()
		key = fmt.Sprintf("leaderboard:weekly:%d-W%02d", year, week)
	default:
		return nil, fmt.Errorf("invalid leaderboard type: %s", leaderboardType)
	}

	// Get data from Redis sorted set
	start := int64(offset)
	stop := int64(offset + limit - 1)
	results, err := s.redis.ZRevRangeWithScores(key, start, stop)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached leaderboard: %v", err)
	}

	var entries []LeaderboardEntry
	for i, result := range results {
		var userData map[string]interface{}
		if err := json.Unmarshal([]byte(result.Member.(string)), &userData); err != nil {
			continue
		}

		entry := LeaderboardEntry{
			UserID:   uint(userData["user_id"].(float64)),
			Username: userData["username"].(string),
			Score:    int64(userData["score"].(float64)),
			Rank:     offset + i + 1,
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
