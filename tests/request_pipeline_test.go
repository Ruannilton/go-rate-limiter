package tests

import (
	"testing"
	"time"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func TestRequestPipeline_LimiterBlocks(t *testing.T) {
	// Limiter with capacity 0 (always blocks)
	limiter := internal.NewFixedWindowRateLimiter(0, 1*time.Second)
	pipeline := internal.NewRequestPipeline(limiter, nil)

	resp, _ := pipeline.HandleRequest()
	if resp.IsAsync() {
		t.Error("Expected sync response when limiter blocks")
	}
	
	if allowed := <-resp.Allowed(); allowed {
		t.Error("Expected request to be blocked")
	}
}

func TestRequestPipeline_LimiterAllows_NoShaper(t *testing.T) {
	limiter := internal.NewFixedWindowRateLimiter(1, 1*time.Second)
	pipeline := internal.NewRequestPipeline(limiter, nil)

	resp, _ := pipeline.HandleRequest()
	// Should be sync (shaper is nil) - actually FixedWindow returns Sync response.
	// But let's check IsAsync() to be sure how pipeline wraps it.
	// Pipeline logic: return limiter (which is sync)
	if resp.IsAsync() {
		t.Error("Expected sync response when no shaper")
	}

	if allowed := <-resp.Allowed(); !allowed {
		t.Error("Expected request to be allowed")
	}
}

func TestRequestPipeline_LimiterAllows_WithShaper(t *testing.T) {
	limiter := internal.NewFixedWindowRateLimiter(1, 1*time.Second)
	
	closeChan := make(chan struct{})
	defer close(closeChan)
	shaper := internal.NewLeakyBucketTrafficShaper(10, 100, closeChan) // Fast shaper

	pipeline := internal.NewRequestPipeline(limiter, shaper)

	resp, _ := pipeline.HandleRequest()

	if !resp.IsAsync() {
		t.Error("Expected async response when shaper is present")
	}

	// Wait for result
	select {
	case allowed := <-resp.Allowed():
		if !allowed {
			t.Error("Expected request to be allowed by shaper")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for shaper")
	}
}
