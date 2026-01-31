package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func TestTokenBucketRateLimiter_Basic(t *testing.T) {
	capacity := 10.0
	refillRate := 1.0 // 1 token per second
	requestCost := 1.0
	limiter := internal.NewTokenBucketRateLimiter(capacity, refillRate, requestCost)

	// Consume all tokens
	for i := 0; i < int(capacity); i++ {
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

	// Wait for 1.1 seconds (should refill ~1.1 tokens -> 1 request)
	time.Sleep(1100 * time.Millisecond)

	resp = limiter.Eval()
	if allowed := <-resp.Allowed(); !allowed {
		t.Errorf("Expected request to be allowed after refill")
	}

	// Should be blocked again
	resp = limiter.Eval()
	if allowed := <-resp.Allowed(); allowed {
		t.Errorf("Expected request to be blocked again")
	}
}

func TestTokenBucketRateLimiter_RefillPrecision(t *testing.T) {
	// Refill 10 tokens/sec. Cost 1.
	capacity := 5.0
	refillRate := 10.0
	requestCost := 1.0
	limiter := internal.NewTokenBucketRateLimiter(capacity, refillRate, requestCost)

	// Consume 5
	for i := 0; i < 5; i++ {
		resp := limiter.Eval()
	<-resp.Allowed()
	}

	// Wait 100ms -> should refill 1 token (10 * 0.1)
	time.Sleep(120 * time.Millisecond) // slightly more to be safe with Milliseconds() truncation

	resp := limiter.Eval()
	if allowed := <-resp.Allowed(); !allowed {
		t.Errorf("Expected request to be allowed after 100ms refill")
	}
}

func TestTokenBucketRateLimiter_Concurrency(t *testing.T) {
	capacity := 50.0
	refillRate := 0.0 // No refill to make counting deterministic
	requestCost := 1.0
	limiter := internal.NewTokenBucketRateLimiter(capacity, refillRate, requestCost)

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

	if allowedCount != int(capacity) {
		t.Errorf("Expected %d allowed requests, got %d", int(capacity), allowedCount)
	}
}
