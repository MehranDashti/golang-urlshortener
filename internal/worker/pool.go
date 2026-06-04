package worker

import (
	"context"
	"log/slog"
	"sync"
)

// Job represents a unit of work.
type Job[T any] struct {
	ID      string
	Payload T
}

// Result holds the outcome of a job.
type Result[T any] struct {
	JobID string
	Value T
	Err   error
}

// Pool is a generic worker pool.
// T = job payload type, R = result type.
type Pool[T any, R any] struct {
	numWorkers int
	jobs       chan Job[T]
	results    chan Result[R]
	process    func(ctx context.Context, job Job[T]) (R, error)
	wg         sync.WaitGroup
}

// NewPool creates a worker pool with numWorkers goroutines.
// process is the function each worker calls for every job.
func NewPool[T any, R any](
	numWorkers int,
	bufferSize int,
	process func(ctx context.Context, job Job[T]) (R, error),
) *Pool[T, R] {
	return &Pool[T, R]{
		numWorkers: numWorkers,
		jobs:       make(chan Job[T], bufferSize),
		results:    make(chan Result[R], bufferSize),
		process:    process,
	}
}

// Start launches the worker goroutines.
// Workers run until the jobs channel is closed.
func (p *Pool[T, R]) Start(ctx context.Context) {
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			slog.Debug("worker started", "id", workerID)

			for job := range p.jobs {
				// Check context before processing
				select {
				case <-ctx.Done():
					slog.Debug("worker stopping — context cancelled",
						"id", workerID)
					return
				default:
					// Context still alive — process the job
				}

				value, err := p.process(ctx, job)
				p.results <- Result[R]{
					JobID: job.ID,
					Value: value,
					Err:   err,
				}
			}

			slog.Debug("worker stopped", "id", workerID)
		}(i)
	}
}

// Submit sends a job to the pool.
// Non-blocking if buffer has space, blocks if full.
func (p *Pool[T, R]) Submit(job Job[T]) {
	p.jobs <- job
}

// TrySubmit sends a job without blocking.
// Returns false if the buffer is full.
func (p *Pool[T, R]) TrySubmit(job Job[T]) bool {
	select {
	case p.jobs <- job:
		return true
	default:
		return false // buffer full — drop the job
	}
}

// Close signals no more jobs and waits for workers to finish.
// Always call Close after you're done submitting jobs.
func (p *Pool[T, R]) Close() {
	close(p.jobs)    // signals workers to exit their range loop
	p.wg.Wait()      // wait for all workers to finish
	close(p.results) // safe to close now — no more writes
}

// Results returns the results channel for reading.
func (p *Pool[T, R]) Results() <-chan Result[R] {
	return p.results
}
