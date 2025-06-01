package services

import (
	"electronic-muyu-backend/internal/models"

	"gorm.io/gorm"
)

type ScoreService struct {
	db          *gorm.DB
	leaderboard *LeaderboardService
	broadcaster *LeaderboardBroadcastService
}

func NewScoreService(db *gorm.DB, leaderboard *LeaderboardService, broadcaster *LeaderboardBroadcastService) *ScoreService {
	return &ScoreService{
		db:          db,
		leaderboard: leaderboard,
		broadcaster: broadcaster,
	}
}

func (s *ScoreService) SubmitScore(userID uint, score int64, gameMode string, duration int) (*models.Score, error) {
	scoreRecord := &models.Score{
		UserID:   userID,
		Score:    score,
		GameMode: gameMode,
		Duration: duration,
	}

	if err := s.db.Create(scoreRecord).Error; err != nil {
		return nil, err
	}

	// 预加载用户信息
	if err := s.db.Preload("User").First(scoreRecord, scoreRecord.ID).Error; err != nil {
		return nil, err
	}

	// 更新Redis缓存（如果可用）
	if s.leaderboard != nil {
		go func() {
			// 在后台更新缓存，不阻塞响应
			s.leaderboard.UpdateLeaderboardCache(userID, scoreRecord.User.Username, score)

			// 广播实时更新（如果可用）
			if s.broadcaster != nil {
				// 获取用户之前的最高分数来计算变化
				var previousHighest int64
				s.db.Model(&models.User{}).Where("id = ?", userID).Select("highest_score").Scan(&previousHighest)

				// 广播分数更新
				s.broadcaster.BroadcastScoreUpdate(userID, scoreRecord.User.Username, score, previousHighest)

				// 如果是新的个人最佳，广播成就
				if score > previousHighest {
					s.broadcaster.BroadcastNewPersonalBest(userID, scoreRecord.User.Username, score)
				}

				// 广播排行榜更新
				s.broadcaster.BroadcastLeaderboardUpdate("global", 10)
			}
		}()
	}

	return scoreRecord, nil
}

func (s *ScoreService) GetUserBestScore(userID uint) (*models.Score, error) {
	var score models.Score
	err := s.db.Where("user_id = ?", userID).
		Order("score DESC").
		First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

func (s *ScoreService) GetUserScoreHistory(userID uint, limit int, offset int) ([]models.Score, error) {
	var scores []models.Score
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&scores).Error
	return scores, err
}

func (s *ScoreService) GetUserTotalScore(userID uint) (int64, error) {
	var total int64
	err := s.db.Model(&models.Score{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(score), 0)").
		Scan(&total).Error
	return total, err
}
