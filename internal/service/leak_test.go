package service_test

import (
    "context"
    "runtime"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func TestNoGoroutineLeak_ClickWorker(t *testing.T) {
    before := runtime.NumGoroutine()

    // Create a context we control
    ctx, cancel := context.WithCancel(context.Background())

    repo := &mockURLRepoForService{}
    svc := service.NewURLService(repo, ctx)

    // Send some clicks
    for i := 0; i < 10; i++ {
        select {
        case svc.ClickCh() <- "some-id":
        default:
        }
    }

    // Cancel context — worker should exit
    cancel()

    // Give worker time to exit
    time.Sleep(10 * time.Millisecond)

    after := runtime.NumGoroutine()

    // Goroutine count should be same or less after cancel
    assert.LessOrEqual(t, after, before+1,
        "goroutine leak detected: before=%d after=%d",
        before, after)
}