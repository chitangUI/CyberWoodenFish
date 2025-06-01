package services

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RateLimiterService struct {
	redisService *RedisService
	logger       *Logger
}

type RateLimit struct {
	Limit   int                 // 请求限制数量
	Window  time.Duration       // 时间窗口
	KeyFunc func(string) string // 生成Redis key的函数
}

func NewRateLimiterService(redisService *RedisService, logger *Logger) *RateLimiterService {
	return &RateLimiterService{
		redisService: redisService,
		logger:       logger,
	}
}

// IsAllowed 检查是否允许请求，使用滑动窗口算法
func (s *RateLimiterService) IsAllowed(identifier string, rateLimit RateLimit) (bool, error) {
	key := rateLimit.KeyFunc(identifier)
	now := time.Now().Unix()
	windowStart := now - int64(rateLimit.Window.Seconds())

	// 使用Redis的事务来原子性地执行操作
	pipe := s.redisService.client.Pipeline()

	// 删除过期的记录
	pipe.ZRemRangeByScore(s.redisService.ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// 计算当前窗口内的请求数
	countCmd := pipe.ZCard(s.redisService.ctx, key)

	// 如果未超限，添加当前请求
	pipe.ZAdd(s.redisService.ctx, key, &redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d-%d", now, time.Now().UnixNano()),
	})

	// 设置key的过期时间
	pipe.Expire(s.redisService.ctx, key, rateLimit.Window)

	_, err := pipe.Exec(s.redisService.ctx)
	if err != nil {
		s.logger.Error("Rate limiter pipeline execution failed - error: %v, key: %s", err, key)
		return false, err
	}

	currentCount := countCmd.Val()

	// 检查是否超过限制（减1是因为我们已经添加了当前请求）
	if currentCount-1 >= int64(rateLimit.Limit) {
		s.logger.Warn("Rate limit exceeded - identifier: %s, key: %s, count: %d, limit: %d",
			identifier, key, currentCount, rateLimit.Limit)

		// 删除刚才添加的请求记录，因为超限了
		s.redisService.client.ZRemRangeByRank(s.redisService.ctx, key, -1, -1)
		return false, nil
	}

	s.logger.Debug("Rate limit check passed - identifier: %s, key: %s, count: %d, limit: %d",
		identifier, key, currentCount, rateLimit.Limit)

	return true, nil
}

// GetRemainingRequests 获取剩余可用请求数
func (s *RateLimiterService) GetRemainingRequests(identifier string, rateLimit RateLimit) (int, error) {
	key := rateLimit.KeyFunc(identifier)
	now := time.Now().Unix()
	windowStart := now - int64(rateLimit.Window.Seconds())

	// 删除过期记录
	err := s.redisService.ZRemRangeByScore(key, "0", fmt.Sprintf("%d", windowStart))
	if err != nil {
		return 0, err
	}

	// 获取当前计数
	currentCount, err := s.redisService.ZCard(key)
	if err != nil {
		return 0, err
	}

	remaining := rateLimit.Limit - int(currentCount)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// Common rate limit configurations
var (
	// 登录限制：每分钟5次
	LoginRateLimit = RateLimit{
		Limit:  5,
		Window: time.Minute,
		KeyFunc: func(identifier string) string {
			return fmt.Sprintf("rate_limit:login:%s", identifier)
		},
	}

	// 注册限制：每小时3次
	RegisterRateLimit = RateLimit{
		Limit:  3,
		Window: time.Hour,
		KeyFunc: func(identifier string) string {
			return fmt.Sprintf("rate_limit:register:%s", identifier)
		},
	}

	// 分数提交限制：每分钟10次
	ScoreSubmitRateLimit = RateLimit{
		Limit:  10,
		Window: time.Minute,
		KeyFunc: func(identifier string) string {
			return fmt.Sprintf("rate_limit:score:%s", identifier)
		},
	}

	// 通用API限制：每分钟100次
	GeneralAPIRateLimit = RateLimit{
		Limit:  100,
		Window: time.Minute,
		KeyFunc: func(identifier string) string {
			return fmt.Sprintf("rate_limit:api:%s", identifier)
		},
	}

	// SSO登录限制：每分钟10次
	SSOLoginRateLimit = RateLimit{
		Limit:  10,
		Window: time.Minute,
		KeyFunc: func(identifier string) string {
			return fmt.Sprintf("rate_limit:sso:%s", identifier)
		},
	}
)
