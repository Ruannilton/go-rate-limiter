package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func TestSlidingWindowLogLimiter_Basic(t *testing.T) {
	capacity := 3
	window := 500 * time.Millisecond
	limiter := internal.NewSlidingWindowLogLimiter(capacity, window)

	// Fill the bucket
	for i := 0; i < capacity; i++ {
		resp := limiter.Eval()
		if allowed := <-resp.Allowed(); !allowed {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}

	// Should be blocked
	resp := limiter.Eval()
	if allowed := <-resp.Allowed(); allowed {
		t.Errorf("Expected request to be blocked")
	}

	// Wait for window to expire
	time.Sleep(window + 50*time.Millisecond)

	// Should be allowed again
	resp = limiter.Eval()
	if allowed := <-resp.Allowed(); !allowed {
		t.Errorf("Expected request to be allowed after window expiration")
	}
}

func TestSlidingWindowLogLimiter_PartialExpiry(t *testing.T) {
	capacity := 2
	window := 200 * time.Millisecond
	limiter := internal.NewSlidingWindowLogLimiter(capacity, window)

	// 1st request
	resp1 := limiter.Eval()
	<-resp1.Allowed()
	
	// Wait half window
	time.Sleep(120 * time.Millisecond)
	
	// 2nd request
	resp2 := limiter.Eval()
	<-resp2.Allowed()

	// 3rd request (blocked)
	resp3 := limiter.Eval()
	if allowed := <-resp3.Allowed(); allowed {
		t.Errorf("Expected request to be blocked")
	}

	// Wait for first request to expire (total > 200ms from start)
	// Current time since start is ~120ms. We need to reach 200ms.
	// Wait another 100ms. Total ~220ms.
	time.Sleep(100 * time.Millisecond)

	// Now first request should be gone, but second is still there.
	// Capacity is 2. Used 1 (the 2nd request).
	// So 1 request should be allowed.
	resp4 := limiter.Eval()
	if allowed := <-resp4.Allowed(); !allowed {
		t.Errorf("Expected request to be allowed after partial expiry")
	}

	// Now we are full again (2nd request + new request).
	resp5 := limiter.Eval()
	if allowed := <-resp5.Allowed(); allowed {
		t.Errorf("Expected request to be blocked again")
	}
}

func TestSlidingWindowLogLimiter_Concurrency(t *testing.T) {
	capacity := 50
	window := 1 * time.Second
	limiter := internal.NewSlidingWindowLogLimiter(capacity, window)

	var wg sync.WaitGroup
	totalRequests := 100
	allowedCount := 0
	var mu sync.Mutex

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := limiter.Eval()
			if allowed := <-resp.Allowed(); allowed {
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