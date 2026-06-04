package util_test

import (
	"sync"
	"testing"
)

// TestRaceCondition_Bad demonstrates a race condition.
// Run with: go test -race ./internal/util/... -run TestRace
// The race detector will catch the unsynchronised counter.
func TestRaceCondition_Bad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping race demonstration in short mode")
	}

	// UNSAFE — demonstrates what NOT to do
	// Uncomment to see the race detector fire:
	//
	// var counter int
	// var wg sync.WaitGroup
	// for i := 0; i < 100; i++ {
	//     wg.Add(1)
	//     go func() {
	//         defer wg.Done()
	//         counter++ // DATA RACE — no lock
	//     }()
	// }
	// wg.Wait()

	// SAFE — correct way with mutex
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()

	if counter != 100 {
		t.Errorf("expected 100, got %d", counter)
	}
}

// TestRaceCondition_Map demonstrates map race.
func TestRaceCondition_Map(t *testing.T) {
	// SAFE — sync.Map handles concurrent access
	var m sync.Map
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		i := i // capture loop variable
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Store(i, i*2)
		}()
	}
	wg.Wait()

	// Verify all values stored correctly
	m.Range(func(k, v interface{}) bool {
		key := k.(int)
		val := v.(int)
		if val != key*2 {
			t.Errorf("expected %d*2=%d, got %d",
				key, key*2, val)
		}
		return true
	})
}
