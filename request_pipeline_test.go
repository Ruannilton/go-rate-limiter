package rate_limiter

import (
	"testing"
	"time"
)

func TestRequestPipeline_LimiterBlocks(t *testing.T) {
	// Limiter with capacity 0 (always blocks)
	limiter := newFixedWindowRateLimiter(0, 1*time.Second)
	pipeline := newRequestPipeline(limiter, nil)

	resp := pipeline.handleRequest()
	if resp.IsAsync() {
		t.Error("Expected sync response when limiter blocks")
	}

	if allowed := <-resp.Allowed(); allowed {
		t.Error("Expected request to be blocked")
	}
}

func TestRequestPipeline_LimiterAllows_NoShaper(t *testing.T) {
	limiter := newFixedWindowRateLimiter(1, 1*time.Second)
	pipeline := newRequestPipeline(limiter, nil)

	resp := pipeline.handleRequest()
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
	limiter := newFixedWindowRateLimiter(1, 1*time.Second)

	closeChan := make(chan struct{})
	defer close(closeChan)
	shaper := newLeakyBucketTrafficShaper(10, 100, closeChan) // Fast shaper

	pipeline := newRequestPipeline(limiter, shaper)

	resp := pipeline.handleRequest()

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
