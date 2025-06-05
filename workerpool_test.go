package workerpool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	pool := NewPool(2)
	defer pool.Close()

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	results := make([]string, 0)

	wg.Add(3)

	for i := range 3 {
		msg := fmt.Sprintf("message%d", i)

		pool.Put((func() {
			defer wg.Done()
			mu.Lock()
			results = append(results, msg)
			mu.Unlock()
		}))
	}

	wg.Wait()

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestWorkerPoolScaling(t *testing.T) {
	pool := NewPool(1)
	defer pool.Close()

	if pool.workersNum != 1 {
		t.Errorf("Expected 1 worker, got %d", pool.Count())
	}

	pool.addWorkers(2)

	if pool.Count() != 3 {
		t.Errorf("Expected 1 worker, got %d", pool.Count())
	}

	pool.Remove(1)

	if pool.Count() != 2 {
		t.Errorf("Expected 2 workers, got %d", pool.Count())
	}

	var wg sync.WaitGroup

	wg.Add(5)

	for range 5 {
		pool.Put(func() {
			time.Sleep(10 * time.Millisecond)
			wg.Done()
		})
	}

	wg.Wait()
}

func TestWorkerPoolClose(t *testing.T) {
	pool := NewPool(2)

	var wg sync.WaitGroup

	wg.Add(2)

	for range 2 {
		pool.Put(func() {
			time.Sleep(50 * time.Millisecond)
			wg.Done()
		})
	}

	done := make(chan struct{})
	go func() {
		pool.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("Close did not wait for the jobs to finish")
	}
}

func TestWorkerPoolStringProcessing(t *testing.T) {
	pool := NewPool(3)
	defer pool.Close()

	type result struct {
		workerID int
		message  string
	}

	var (
		results []result
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	wg.Add(5)

	for i := range 5 {
		msg := fmt.Sprintf("test%d", i)

		pool.Put(func() {
			defer wg.Done()
			mu.Lock()

			workerID := i + 1
			results = append(results, result{
				workerID: workerID,
				message:  msg,
			})
			mu.Unlock()
		})
	}

	wg.Wait()

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}
