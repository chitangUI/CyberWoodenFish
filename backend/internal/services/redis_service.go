package services

import (
	"context"
	"electronic-muyu-backend/internal/config"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisService(cfg *config.Config) (*RedisService, error) {
	if cfg.RedisURL == "" {
		return nil, nil // Redis 是可选的
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	// 测试连接
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisService{
		client: client,
		ctx:    ctx,
	}, nil
}

// Set stores a key-value pair with expiration
func (r *RedisService) Set(key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, string(jsonValue), expiration).Err()
}

// Get retrieves a value by key
func (r *RedisService) Get(key string, dest interface{}) error {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key
func (r *RedisService) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Exists checks if a key exists
func (r *RedisService) Exists(key string) (bool, error) {
	count, err := r.client.Exists(r.ctx, key).Result()
	return count > 0, err
}

// Increment increments a key's value
func (r *RedisService) Increment(key string) (int64, error) {
	return r.client.Incr(r.ctx, key).Result()
}

// SetNX sets a key only if it doesn't exist (used for locking)
func (r *RedisService) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	return r.client.SetNX(r.ctx, key, string(jsonValue), expiration).Result()
}

// ZAdd adds a member to a sorted set
func (r *RedisService) ZAdd(key string, score float64, member interface{}) error {
	jsonMember, err := json.Marshal(member)
	if err != nil {
		return err
	}
	z := &redis.Z{
		Score:  score,
		Member: string(jsonMember),
	}
	return r.client.ZAdd(r.ctx, key, z).Err()
}

// ZRevRange gets members from a sorted set in descending order
func (r *RedisService) ZRevRange(key string, start, stop int64) ([]string, error) {
	return r.client.ZRevRange(r.ctx, key, start, stop).Result()
}

// ZRevRangeWithScores gets members with scores from a sorted set in descending order
func (r *RedisService) ZRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	return r.client.ZRevRangeWithScores(r.ctx, key, start, stop).Result()
}

// ZRank gets the rank of a member in a sorted set
func (r *RedisService) ZRank(key string, member interface{}) (int64, error) {
	jsonMember, err := json.Marshal(member)
	if err != nil {
		return 0, err
	}
	return r.client.ZRank(r.ctx, key, string(jsonMember)).Result()
}

// ZRemRangeByScore removes members by score range
func (r *RedisService) ZRemRangeByScore(key, min, max string) error {
	return r.client.ZRemRangeByScore(r.ctx, key, min, max).Err()
}

// ZCard gets the number of members in a sorted set
func (r *RedisService) ZCard(key string) (int64, error) {
	return r.client.ZCard(r.ctx, key).Result()
}

// ZRemRangeByRank removes members by rank range
func (r *RedisService) ZRemRangeByRank(key string, start, stop int64) error {
	return r.client.ZRemRangeByRank(r.ctx, key, start, stop).Err()
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	return r.client.Close()
}

// Publish publishes a message to a channel
func (r *RedisService) Publish(channel string, message interface{}) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return r.client.Publish(r.ctx, channel, string(jsonMessage)).Err()
}

// Subscribe subscribes to a channel
func (r *RedisService) Subscribe(channel string) *redis.PubSub {
	return r.client.Subscribe(r.ctx, channel)
}
