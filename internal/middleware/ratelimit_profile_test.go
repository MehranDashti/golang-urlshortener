package middleware_test

import (
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"urlshortener/internal/middleware"
)

func TestProfileRateLimiter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping profile test in short mode")
	}

	rl := middleware.NewRateLimiter(1000, time.Minute)

	// Start CPU profile
	f, err := os.Create("ratelimit_cpu.prof")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("failed to close cpu profile file: %v", err)
		}
	}()

	if err := pprof.StartCPUProfile(f); err != nil {
		t.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// Run the code to profile
	for i := 0; i < 100000; i++ {
		rl.Allow("192.0.2.1")
	}

	t.Log("profile written to ratelimit_cpu.prof")
	t.Log("analyse with: go tool pprof ratelimit_cpu.prof")
}

func TestProfileMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory profile test in short mode")
	}

	rl := middleware.NewRateLimiter(1000, time.Minute)

	// Generate load with many different IPs
	for i := 0; i < 10000; i++ {
		ip := "192.0.2." + string(rune('0'+i%255))
		rl.Allow(ip)
	}

	// Write heap profile
	f, err := os.Create("ratelimit_heap.prof")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("failed to close heap profile file: %v", err)
		}
	}()

	if err := pprof.WriteHeapProfile(f); err != nil {
		t.Error(err)
	}
	t.Log("heap profile written to ratelimit_heap.prof")
	t.Log("analyse with: go tool pprof ratelimit_heap.prof")
}
