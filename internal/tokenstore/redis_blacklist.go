package tokenstore

import (
	"context"
	"fmt"
	"time"

	"urlshortener/internal/cache"
)

const blacklistPrefix = "blacklist:"

// RedisBlacklist stores revoked JWT IDs in Redis with automatic TTL expiry.
// When a token's TTL expires, Redis removes the key automatically —
// no cleanup goroutine needed, unlike the sync.Map implementation.
type RedisBlacklist struct {
	cache *cache.RedisCache
}

func NewRedisBlacklist(c *cache.RedisCache) *RedisBlacklist {
	return &RedisBlacklist{cache: c}
}

// Revoke adds a token JTI to the blacklist until its expiry time.
func (b *RedisBlacklist) Revoke(ctx context.Context, jti string, expiry time.Duration) error {
	key := fmt.Sprintf("%s%s", blacklistPrefix, jti)
	return b.cache.Set(ctx, key, []byte("1"), expiry)
}

// IsRevoked returns true if the token JTI has been blacklisted.
func (b *RedisBlacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("%s%s", blacklistPrefix, jti)
	return b.cache.Exists(ctx, key)
}
