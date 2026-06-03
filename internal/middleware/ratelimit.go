package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
)

type visitor struct {
    timestamps []time.Time
    mu         sync.Mutex // per-visitor lock stays Mutex — always written
}

type RateLimiter struct {
    visitors map[string]*visitor
    mu       sync.RWMutex  // ← changed from sync.Mutex
    limit    int
    window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    rl := &RateLimiter{
        visitors: make(map[string]*visitor),
        limit:    limit,
        window:   window,
    }
    go rl.cleanupLoop()
    return rl
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
    // First try read lock — cheap, allows concurrent reads
    rl.mu.RLock()
    v, exists := rl.visitors[ip]
    rl.mu.RUnlock()

    if exists {
        return v // fast path — no write needed
    }

    // Visitor doesn't exist — need write lock
    rl.mu.Lock()
    defer rl.mu.Unlock()

    // Check again after acquiring write lock
    // Another goroutine may have created it between RUnlock and Lock
    // This is called "double-checked locking"
    if v, exists = rl.visitors[ip]; exists {
        return v
    }

    v = &visitor{}
    rl.visitors[ip] = v
    return v
}

func (rl *RateLimiter) allow(ip string) bool {
    v := rl.getVisitor(ip)

    v.mu.Lock()
    defer v.mu.Unlock()

    now := time.Now()
    windowStart := now.Add(-rl.window)

    valid := v.timestamps[:0]
    for _, t := range v.timestamps {
        if t.After(windowStart) {
            valid = append(valid, t)
        }
    }
    v.timestamps = valid

    if len(v.timestamps) >= rl.limit {
        return false
    }

    v.timestamps = append(v.timestamps, now)
    return true
}

func (rl *RateLimiter) cleanupLoop() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        rl.mu.Lock() // full write lock for cleanup
        now := time.Now()
        for ip, v := range rl.visitors {
            v.mu.Lock()
            if len(v.timestamps) == 0 ||
                v.timestamps[len(v.timestamps)-1].
                    Before(now.Add(-rl.window)) {
                delete(rl.visitors, ip)
            }
            v.mu.Unlock()
        }
        rl.mu.Unlock()
    }
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        if !rl.allow(ip) {
            c.JSON(http.StatusTooManyRequests,
                map[string]interface{}{
                    "success": false,
                    "code":    429,
                    "message": "too many requests — slow down",
                })
            c.Abort()
            return
        }
        c.Next()
    }
}

// Allow is exported for benchmarks and testing.
func (rl *RateLimiter) Allow(ip string) bool {
    return rl.allow(ip)
}