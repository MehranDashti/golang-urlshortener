package tokenstore

import (
	"context"
	"sync"
	"time"
)

type TokenBlacklist interface {
	Revoke(ctx context.Context, jti string, expiry time.Duration) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
}

// Blacklist stores revoked token JTIs in memory.
// For production: replace with Redis using SETEX for automatic expiry.
type Blacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // jti → expiry time
}

func NewBlacklist() *Blacklist {
	b := &Blacklist{
		tokens: make(map[string]time.Time),
	}
	go b.cleanupLoop() // remove expired entries
	return b
}

func (b *Blacklist) Revoke(_ context.Context, jti string, expiry time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens[jti] = time.Now().Add(expiry)
	return nil
}

func (b *Blacklist) IsRevoked(_ context.Context, jti string) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	expiry, exists := b.tokens[jti]
	if !exists {
		return false, nil
	}
	return time.Now().Before(expiry), nil
}

// cleanupLoop removes expired entries every minute
func (b *Blacklist) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()
		now := time.Now()
		for jti, expiry := range b.tokens {
			if now.After(expiry) {
				delete(b.tokens, jti)
			}
		}
		b.mu.Unlock()
	}
}
