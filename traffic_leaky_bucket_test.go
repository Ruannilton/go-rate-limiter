package rate_limiter

import (
	"testing"
	"time"
)

func TestLeakyBucketTrafficShaper_Basic(t *testing.T) {
	capacity := 10
	rate := 10 // 10 per second -> 100ms interval
	closeChan := make(chan struct{})
	defer close(closeChan)

	shaper := newLeakyBucketTrafficShaper(capacity, rate, closeChan)

	// Add a request
	start := time.Now()
	respChan := shaper.addRequest()

	// Wait for response
	select {
	case val := <-respChan:
		if !val {
			t.Errorf("Expected true from shaper")
		}
		// First request might be immediate or wait for next tick?
		// NewTicker fires after interval. So it should wait ~100ms.
		if time.Since(start) < 50*time.Millisecond {
			t.Logf("Warning: Response was too fast, ticker might behave differently depending on implementation detail")
		}
	case <-time.After(200 * time.Millisecond):
		t.Errorf("Timed out waiting for shaper response")
	}
}

func TestLeakyBucketTrafficShaper_Queue(t *testing.T) {
	capacity := 2
	rate := 20 // 50ms
	closeChan := make(chan struct{})
	defer close(closeChan)

	shaper := newLeakyBucketTrafficShaper(capacity, rate, closeChan)

	// Add 2 requests (fits in queue)
	ch1 := shaper.addRequest()
	ch2 := shaper.addRequest()

	// Both should eventually return
	timeout := time.After(1 * time.Second)

	select {
	case <-ch1:
	case <-timeout:
		t.Fatal("Timeout waiting for ch1")
	}

	select {
	case <-ch2:
	case <-timeout:
		t.Fatal("Timeout waiting for ch2")
	}
}

func TestLeakyBucketTrafficShaper_Blocking(t *testing.T) {
	// Capacity 1. Rate very slow (1 per second)
	capacity := 1
	rate := 1
	closeChan := make(chan struct{})
	defer close(closeChan)

	shaper := newLeakyBucketTrafficShaper(capacity, rate, closeChan)

	// Fill queue
	ch1 := shaper.addRequest()

	// Next add should block until ticker fires (1s) and frees space
	done := make(chan struct{})
	go func() {
		shaper.addRequest()
		close(done)
	}()

	select {
	case <-done:
		t.Errorf("AddRequest should have blocked (queue full)")
	case <-time.After(100 * time.Millisecond):
		// Correct behavior: blocked for at least 100ms
	}

	// Eventually it should finish (after ~1s)
	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Errorf("Timed out waiting for blocked AddRequest to finish")
	}

	// Consume ch1 to be clean
	<-ch1
}
