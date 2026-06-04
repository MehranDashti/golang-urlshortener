package tokenstore

import (
	"sync"
	"time"
)

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

// Revoke adds a JTI to the blacklist until its expiry.
func (b *Blacklist) Revoke(jti string, expiry time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens[jti] = expiry
}

// IsRevoked checks if a JTI has been revoked.
func (b *Blacklist) IsRevoked(jti string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	expiry, exists := b.tokens[jti]
	if !exists {
		return false
	}
	// If expired — treat as not revoked (token would fail JWT validation anyway)
	return time.Now().Before(expiry)
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
