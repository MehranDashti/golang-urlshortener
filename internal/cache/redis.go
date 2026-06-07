package cache

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

// ErrCacheMiss is returned when a key doesn't exist in the cache.
// Callers use errors.Is(err, cache.ErrCacheMiss) — same sentinel pattern
// you used in the repository layer.
var ErrCacheMiss = errors.New("cache miss")

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr, password string, db int) *RedisCache {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
    return &RedisCache{client: client}
}

func (c *RedisCache) Ping(ctx context.Context) error {
    return c.client.Ping(ctx).Err()
}

// Set stores a value with a TTL.
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
        return fmt.Errorf("cache set %s: %w", key, err)
    }
    return nil
}

// Get retrieves a value. Returns ErrCacheMiss if the key doesn't exist.
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
    val, err := c.client.Get(ctx, key).Bytes()
    if errors.Is(err, redis.Nil) {
        return nil, ErrCacheMiss  // redis.Nil = key not found
    }
    if err != nil {
        return nil, fmt.Errorf("cache get %s: %w", key, err)
    }
    return val, nil
}

// Delete removes a key.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
    if err := c.client.Del(ctx, key).Err(); err != nil {
        return fmt.Errorf("cache delete %s: %w", key, err)
    }
    return nil
}

// Exists checks if a key exists.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
    n, err := c.client.Exists(ctx, key).Result()
    if err != nil {
        return false, fmt.Errorf("cache exists %s: %w", key, err)
    }
    return n > 0, nil
}