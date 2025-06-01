package services

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type LeaderboardBroadcastService struct {
	redisService       *RedisService
	leaderboardService *LeaderboardService
	logger             *Logger
}

type LeaderboardUpdate struct {
	Type      string      `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type ScoreUpdate struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	NewScore     int64  `json:"new_score"`
	PreviousRank int    `json:"previous_rank"`
	NewRank      int    `json:"new_rank"`
	ScoreChange  int64  `json:"score_change"`
}

type RankingChange struct {
	TopEntries []LeaderboardEntry `json:"top_entries"`
	Limit      int                `json:"limit"`
	UpdateType string             `json:"update_type"` // "global", "daily", "weekly"
}

func NewLeaderboardBroadcastService(
	redisService *RedisService,
	leaderboardService *LeaderboardService,
	logger *Logger,
) *LeaderboardBroadcastService {
	return &LeaderboardBroadcastService{
		redisService:       redisService,
		leaderboardService: leaderboardService,
		logger:             logger,
	}
}

// BroadcastScoreUpdate 广播分数更新
func (s *LeaderboardBroadcastService) BroadcastScoreUpdate(userID uint, username string, newScore, previousScore int64) error {
	// 获取用户在排行榜中的新排名
	newRank, err := s.getUserRank(userID, "global")
	if err != nil {
		s.logger.Error("Failed to get user rank: %v", err)
		newRank = 0
	}

	// 构建分数更新消息
	scoreUpdate := ScoreUpdate{
		UserID:      userID,
		Username:    username,
		NewScore:    newScore,
		NewRank:     newRank,
		ScoreChange: newScore - previousScore,
	}

	update := LeaderboardUpdate{
		Type:      "score_update",
		Timestamp: getCurrentTimestamp(),
		Data:      scoreUpdate,
	}

	// 广播到排行榜频道
	return s.broadcastToChannel("leaderboard:updates", update)
}

// BroadcastLeaderboardUpdate 广播排行榜更新
func (s *LeaderboardBroadcastService) BroadcastLeaderboardUpdate(updateType string, limit int) error {
	var entries []LeaderboardEntry
	var err error

	switch updateType {
	case "global":
		entries, err = s.leaderboardService.GetGlobalLeaderboard(limit, 0)
	case "daily":
		entries, err = s.leaderboardService.GetDailyLeaderboard(limit, 0)
	case "weekly":
		entries, err = s.leaderboardService.GetWeeklyLeaderboard(limit, 0)
	default:
		return fmt.Errorf("unknown update type: %s", updateType)
	}

	if err != nil {
		return fmt.Errorf("failed to get leaderboard: %v", err)
	}

	rankingChange := RankingChange{
		TopEntries: entries,
		Limit:      limit,
		UpdateType: updateType,
	}

	update := LeaderboardUpdate{
		Type:      "leaderboard_update",
		Timestamp: getCurrentTimestamp(),
		Data:      rankingChange,
	}

	// 广播到相应的频道
	channelName := fmt.Sprintf("leaderboard:%s", updateType)
	return s.broadcastToChannel(channelName, update)
}

// BroadcastNewPersonalBest 广播新的个人最佳成绩
func (s *LeaderboardBroadcastService) BroadcastNewPersonalBest(userID uint, username string, newScore int64) error {
	personalBest := map[string]interface{}{
		"user_id":   userID,
		"username":  username,
		"new_score": newScore,
		"message":   fmt.Sprintf("%s achieved a new personal best: %d points!", username, newScore),
	}

	update := LeaderboardUpdate{
		Type:      "personal_best",
		Timestamp: getCurrentTimestamp(),
		Data:      personalBest,
	}

	return s.broadcastToChannel("leaderboard:achievements", update)
}

// BroadcastRankChange 广播排名变化
func (s *LeaderboardBroadcastService) BroadcastRankChange(userID uint, username string, oldRank, newRank int, score int64) error {
	if oldRank == newRank {
		return nil // 排名未变化，无需广播
	}

	rankChange := map[string]interface{}{
		"user_id":   userID,
		"username":  username,
		"old_rank":  oldRank,
		"new_rank":  newRank,
		"score":     score,
		"direction": getRankDirection(oldRank, newRank),
	}

	update := LeaderboardUpdate{
		Type:      "rank_change",
		Timestamp: getCurrentTimestamp(),
		Data:      rankChange,
	}

	return s.broadcastToChannel("leaderboard:rank_changes", update)
}

// SubscribeToLeaderboardUpdates 订阅排行榜更新
func (s *LeaderboardBroadcastService) SubscribeToLeaderboardUpdates(channels []string) (*RedisPubSub, error) {
	if s.redisService == nil {
		return nil, fmt.Errorf("redis service not available")
	}

	// 创建订阅
	pubsub := &RedisPubSub{
		channels: channels,
		pubsub:   s.redisService.client.Subscribe(s.redisService.ctx, channels...),
	}

	return pubsub, nil
}

// 私有辅助方法

func (s *LeaderboardBroadcastService) broadcastToChannel(channel string, update LeaderboardUpdate) error {
	if s.redisService == nil {
		s.logger.Debug("Redis not available, skipping broadcast to channel: %s", channel)
		return nil
	}

	return s.redisService.Publish(channel, update)
}

func (s *LeaderboardBroadcastService) getUserRank(userID uint, leaderboardType string) (int, error) {
	// 这里可以实现具体的排名查询逻辑
	// 暂时返回0，实际实现需要查询Redis或数据库
	return 0, nil
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func getRankDirection(oldRank, newRank int) string {
	if oldRank == 0 {
		return "new"
	}
	if newRank < oldRank {
		return "up"
	}
	return "down"
}

// RedisPubSub Redis发布订阅包装器
type RedisPubSub struct {
	channels []string
	pubsub   *redis.PubSub
}

func (p *RedisPubSub) GetChannel() <-chan *redis.Message {
	return p.pubsub.Channel()
}

func (p *RedisPubSub) Close() error {
	return p.pubsub.Close()
}
