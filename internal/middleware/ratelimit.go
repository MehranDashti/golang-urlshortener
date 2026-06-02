package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
)

// visitor tracks request timestamps for one IP
type visitor struct {
    timestamps []time.Time
    mu         sync.Mutex // each visitor has their own lock
}

// RateLimiter holds all visitors and its own lock
type RateLimiter struct {
    visitors map[string]*visitor
    mu       sync.Mutex // protects the visitors map
    limit    int
    window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    rl := &RateLimiter{
        visitors: make(map[string]*visitor),
        limit:    limit,
        window:   window,
    }

    // Background goroutine — clean up stale visitors every minute
    // Prevents memory leak from IPs that stop sending requests
    go rl.cleanupLoop()

    return rl
}

// getVisitor returns the visitor for an IP, creating one if needed
func (rl *RateLimiter) getVisitor(ip string) *visitor {
    rl.mu.Lock()
    defer rl.mu.Unlock() // ← defer in action — unlocks when function returns

    v, exists := rl.visitors[ip]
    if !exists {
        v = &visitor{}
        rl.visitors[ip] = v
    }
    return v
}

// allow checks and records a request for the given IP
func (rl *RateLimiter) allow(ip string) bool {
    v := rl.getVisitor(ip)

    v.mu.Lock()
    defer v.mu.Unlock()

    now := time.Now()
    windowStart := now.Add(-rl.window)

    // Remove timestamps outside the window — sliding window
    valid := v.timestamps[:0] // reuse the slice memory
    for _, t := range v.timestamps {
        if t.After(windowStart) {
            valid = append(valid, t)
        }
    }
    v.timestamps = valid

    // Check limit
    if len(v.timestamps) >= rl.limit {
        return false // too many requests
    }

    // Record this request
    v.timestamps = append(v.timestamps, now)
    return true
}

// cleanupLoop removes visitors with no recent requests
// Runs as a background goroutine — prevents unbounded memory growth
func (rl *RateLimiter) cleanupLoop() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for range ticker.C { // blocks until ticker fires
        rl.mu.Lock()
        now := time.Now()
        for ip, v := range rl.visitors {
            v.mu.Lock()
            if len(v.timestamps) == 0 ||
                v.timestamps[len(v.timestamps)-1].Before(
                    now.Add(-rl.window)) {
                delete(rl.visitors, ip)
            }
            v.mu.Unlock()
        }
        rl.mu.Unlock()
    }
}

// Middleware returns a Gin middleware function
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()

        if !rl.allow(ip) {
            c.JSON(http.StatusTooManyRequests, map[string]interface{}{
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