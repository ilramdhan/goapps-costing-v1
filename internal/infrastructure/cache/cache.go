package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/homindolenern/goapps-costing-v1/internal/infrastructure/redis"
)

// Cache provides a generic caching interface.
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) (bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeleteByPattern(ctx context.Context, pattern string) error
}

// RedisCache implements Cache using Redis.
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis-backed cache.
func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	if client == nil {
		return nil
	}
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// Get retrieves a value from cache.
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	if c == nil || c.client == nil {
		return false, nil
	}

	err := c.client.Get(ctx, c.prefix+key, dest)
	if err != nil {
		// Key not found is not an error, just cache miss.
		// We intentionally ignore the error here since Redis returns
		// an error for cache misses, which is expected behavior.
		return false, nil //nolint:nilerr // cache miss is expected
	}
	return true, nil
}

// Set stores a value in cache.
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, c.prefix+key, value, ttl)
}

// Delete removes keys from cache.
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if c == nil || c.client == nil {
		return nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = c.prefix + k
	}
	return c.client.Delete(ctx, prefixedKeys...)
}

// DeleteByPattern removes keys matching a pattern.
func (c *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.DeleteByPattern(ctx, c.prefix+pattern)
}

// NoOpCache is a cache that does nothing (for when Redis is not available).
type NoOpCache struct{}

// NewNoOpCache creates a no-op cache.
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (c *NoOpCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	return false, nil
}

func (c *NoOpCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (c *NoOpCache) DeleteByPattern(ctx context.Context, pattern string) error {
	return nil
}

// Cached wraps a function with caching.
func Cached[T any](
	ctx context.Context,
	cache Cache,
	key string,
	ttl time.Duration,
	fn func() (T, error),
) (T, error) {
	var result T

	// Try to get from cache
	found, err := cache.Get(ctx, key, &result)
	if err == nil && found {
		return result, nil
	}

	// Execute function
	result, err = fn()
	if err != nil {
		return result, err
	}

	// Cache the result
	_ = cache.Set(ctx, key, result, ttl)

	return result, nil
}

// Key generates a cache key from components.
func Key(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}

	key := parts[0]
	for _, p := range parts[1:] {
		key += ":" + p
	}
	return key
}

// MarshalJSON is a helper to serialize for caching.
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalJSON is a helper to deserialize from cache.
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
