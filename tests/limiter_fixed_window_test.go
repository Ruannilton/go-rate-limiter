package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func TestFixedWindowRateLimiter_Basic(t *testing.T) {
	capacity := 2
	interval := 200 * time.Millisecond
	limiter := internal.NewFixedWindowRateLimiter(capacity, interval)

	// First request should pass
	if resp := limiter.Eval(); !<-resp.Allowed() {
		t.Errorf("Expected first request to be allowed")
	}

	// Second request should pass
	if resp := limiter.Eval(); !<-resp.Allowed() {
		t.Errorf("Expected second request to be allowed")
	}

	// Third request should be blocked
	if resp := limiter.Eval(); <-resp.Allowed() {
		t.Errorf("Expected third request to be blocked")
	}

	// Wait for window reset
	time.Sleep(interval + 50*time.Millisecond)

	// Should be allowed again
	if resp := limiter.Eval(); !<-resp.Allowed() {
		t.Errorf("Expected request to be allowed after reset")
	}
}

func TestFixedWindowRateLimiter_Concurrency(t *testing.T) {
	capacity := 50
	interval := 1 * time.Second
	limiter := internal.NewFixedWindowRateLimiter(capacity, interval)

	var wg sync.WaitGroup
	totalRequests := 100
	allowedCount := 0
	var mu sync.Mutex

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := limiter.Eval()
			if <-resp.Allowed() {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowedCount != capacity {
		t.Errorf("Expected %d allowed requests, got %d", capacity, allowedCount)
	}
}
