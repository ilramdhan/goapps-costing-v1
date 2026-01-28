package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/homindolenern/goapps-costing-v1/internal/config"
)

// Client wraps the Redis client
type Client struct {
	rdb *redis.Client
}

// NewClient creates a new Redis client
func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Int("db", cfg.DB).
		Msg("Connected to Redis")

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}

// HealthCheck verifies the Redis connection is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Set stores a value with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return c.rdb.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete removes a key
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Keys returns keys matching a pattern
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.rdb.Keys(ctx, pattern).Result()
}

// DeleteByPattern deletes all keys matching a pattern
func (c *Client) DeleteByPattern(ctx context.Context, pattern string) error {
	keys, err := c.Keys(ctx, pattern)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.Delete(ctx, keys...)
	}
	return nil
}

// Cache key prefixes
const (
	UOMKeyPrefix       = "uom:"
	ParameterKeyPrefix = "param:"
)

// UOM cache keys
func UOMCacheKey(code string) string {
	return UOMKeyPrefix + code
}

func UOMListCacheKey(page, pageSize int, category string) string {
	return fmt.Sprintf("%slist:%d:%d:%s", UOMKeyPrefix, page, pageSize, category)
}

// Parameter cache keys
func ParameterCacheKey(code string) string {
	return ParameterKeyPrefix + code
}

func ParameterListCacheKey(page, pageSize int, category string, isActive *bool) string {
	activeStr := "all"
	if isActive != nil {
		if *isActive {
			activeStr = "active"
		} else {
			activeStr = "inactive"
		}
	}
	return fmt.Sprintf("%slist:%d:%d:%s:%s", ParameterKeyPrefix, page, pageSize, category, activeStr)
}
