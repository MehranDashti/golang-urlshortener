package middleware_test

import (
    "fmt"
    "testing"
    "time"

    "urlshortener/internal/middleware"
)

func BenchmarkRateLimiter_Allow(b *testing.B) {
    rl := middleware.NewRateLimiter(1000, time.Minute)
    b.ReportAllocs()
    b.ResetTimer() // don't count setup time

    for i := 0; i < b.N; i++ {
        rl.Allow("192.0.2.1")
    }
}

// BenchmarkRateLimiter_Allow_ManyIPs tests with different IPs
// — exercises the map lookup and visitor creation path.
func BenchmarkRateLimiter_Allow_ManyIPs(b *testing.B) {
    rl := middleware.NewRateLimiter(1000, time.Minute)
    b.ReportAllocs()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        ip := fmt.Sprintf("192.0.2.%d", i%255)
        rl.Allow(ip)
    }
}

// BenchmarkRateLimiter_Parallel tests concurrent access.
func BenchmarkRateLimiter_Parallel(b *testing.B) {
    rl := middleware.NewRateLimiter(1000, time.Minute)
    b.ReportAllocs()
    b.ResetTimer()

    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            ip := fmt.Sprintf("192.0.2.%d", i%255)
            rl.Allow(ip)
            i++
        }
    })
}