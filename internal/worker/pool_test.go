package worker_test

import (
    "context"
    "fmt"
    "sync/atomic"
    "testing"

    "github.com/stretchr/testify/assert"
    "urlshortener/internal/worker"
)

func TestPool_ProcessesAllJobs(t *testing.T) {
    const numJobs    = 100
    const numWorkers = 5

    var processed atomic.Int64 // atomic counter — safe for concurrent use

    pool := worker.NewPool[int, int](
        numWorkers,
        numJobs,
        func(ctx context.Context,
            job worker.Job[int]) (int, error) {
            processed.Add(1)
            return job.Payload * 2, nil
        },
    )

    pool.Start(context.Background())

    for i := 0; i < numJobs; i++ {
        pool.Submit(worker.Job[int]{
            ID:      fmt.Sprintf("job-%d", i),
            Payload: i,
        })
    }

    go pool.Close()

    results := make([]int, 0, numJobs)
    for r := range pool.Results() {
        assert.NoError(t, r.Err)
        results = append(results, r.Value)
    }

    assert.Equal(t, int64(numJobs), processed.Load())
    assert.Len(t, results, numJobs)
}

func TestPool_HandlesErrors(t *testing.T) {
    pool := worker.NewPool[string, string](
        2, 10,
        func(ctx context.Context,
            job worker.Job[string]) (string, error) {
            if job.Payload == "bad" {
                return "", fmt.Errorf("bad job: %s", job.ID)
            }
            return job.Payload + "_ok", nil
        },
    )

    pool.Start(context.Background())
    pool.Submit(worker.Job[string]{ID: "1", Payload: "good"})
    pool.Submit(worker.Job[string]{ID: "2", Payload: "bad"})
    pool.Submit(worker.Job[string]{ID: "3", Payload: "good"})

    go pool.Close()

    var errCount int
    for r := range pool.Results() {
        if r.Err != nil {
            errCount++
        }
    }

    assert.Equal(t, 1, errCount)
}

func TestPool_ContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())

    var processed atomic.Int64

    pool := worker.NewPool[int, int](
        3, 100,
        func(ctx context.Context,
            job worker.Job[int]) (int, error) {
            processed.Add(1)
            return job.Payload, nil
        },
    )

    pool.Start(ctx)

    // Submit jobs
    for i := 0; i < 100; i++ {
        pool.Submit(worker.Job[int]{
            ID:      fmt.Sprintf("job-%d", i),
            Payload: i,
        })
    }

    // Cancel context early — workers should stop
    cancel()
    go pool.Close()

    // Drain results
    for range pool.Results() {}

    // Not all jobs processed — context was cancelled
    t.Logf("processed %d/100 jobs before cancellation",
        processed.Load())
}

// BenchmarkPool measures throughput
func BenchmarkPool(b *testing.B) {
    pool := worker.NewPool[int, int](
        10, 1000,
        func(ctx context.Context,
            job worker.Job[int]) (int, error) {
            return job.Payload * 2, nil
        },
    )

    pool.Start(context.Background())
    b.ResetTimer()

    go func() {
        for i := 0; i < b.N; i++ {
            pool.Submit(worker.Job[int]{
                ID:      fmt.Sprintf("job-%d", i),
                Payload: i,
            })
        }
        pool.Close()
    }()

    for range pool.Results() {}
}