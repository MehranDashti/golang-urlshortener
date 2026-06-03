package util_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "urlshortener/internal/util"
)

func TestGenerateShortCode(t *testing.T) {
    tests := []struct {
        name  string
        check func(t *testing.T, code string)
    }{
        {
            name: "correct length",
            check: func(t *testing.T, code string) {
                assert.Len(t, code, 6)
            },
        },
        {
            name: "valid characters only",
            check: func(t *testing.T, code string) {
                charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
                for _, c := range code {
                    assert.Contains(t, charset, string(c))
                }
            },
        },
        {
            name: "not empty",
            check: func(t *testing.T, code string) {
                assert.NotEmpty(t, code)
            },
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            code := util.GenerateShortCode()
            tt.check(t, code)
        })
    }
}

func TestGenerateShortCode_Uniqueness(t *testing.T) {
    // Generate many codes and check for collisions
    const count = 1000
    seen := make(map[string]bool, count)

    for i := 0; i < count; i++ {
        code := util.GenerateShortCode()
        assert.False(t, seen[code],
            "collision detected: %s generated twice", code)
        seen[code] = true
    }
}

// BenchmarkGenerateShortCode measures raw generation speed.
// Run: go test -bench=BenchmarkGenerate -benchmem ./internal/util/...
func BenchmarkGenerateShortCode(b *testing.B) {
    b.ReportAllocs() // report memory allocations
    for i := 0; i < b.N; i++ {
        util.GenerateShortCode()
    }
}

// BenchmarkGenerateShortCode_Parallel measures concurrent performance.
// Simulates multiple goroutines generating codes simultaneously.
func BenchmarkGenerateShortCode_Parallel(b *testing.B) {
    b.ReportAllocs()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            util.GenerateShortCode()
        }
    })
}