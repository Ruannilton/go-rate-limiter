package tests

import (
	"testing"
	"time"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func TestSlidingWindowLogAlgorithm(t *testing.T) {
	storage := internal.NewSlidingWindowLogMemoryStorage()
	identifier := "test-sliding-window-log"
	capacity := 3
	windowSize := 1 * time.Second

	storage.SetDefaultValue(identifier, internal.SlidingWindowLogCtrParams{Capacity: capacity, WindowSize: windowSize})
	handler := internal.NewHandler(storage)

	// Allow requests up to capacity within the window
	for i := 0; i < capacity; i++ {
		response, err := handler.Handle(identifier)
		if err != nil {
			t.Fatalf("Unexpected error on request %d: %v", i+1, err)
		}
		if !response.IsAllowed() {
			t.Errorf("Request %d should have been allowed, but was denied", i+1)
		}
		time.Sleep(50 * time.Millisecond) // Simulate some time passing
	}

	// Deny next request as capacity is full
	response, err := handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error on denied request: %v", err)
	}
	if response.IsAllowed() {
		t.Error("Request should have been denied, but was allowed (capacity exceeded)")
	}

	// Wait for the window to pass
	time.Sleep(windowSize + 50*time.Millisecond)

	// Allow requests again as old logs should be gone
	// The next 'capacity' requests should be allowed.
	for i := 0; i < capacity; i++ {
		response, err := handler.Handle(identifier)
		if err != nil {
			t.Fatalf("Unexpected error on request %d after window passed: %v", i+1, err)
		}
		if !response.IsAllowed() {
			t.Errorf("Request %d should have been allowed after window passed, but was denied", i+1)
		}
		time.Sleep(50 * time.Millisecond) // Simulate some time passing
	}

	// Deny next request as capacity is full again
	response, err = handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error on denied request after window passed: %v", err)
	}
	if response.IsAllowed() {
		t.Error("Request should have been denied after window passed (capacity exceeded again), but was allowed")
	}
}

func TestSlidingWindowLogAlgorithm_WindowClears(t *testing.T) {
	storage := internal.NewSlidingWindowLogMemoryStorage()
	identifier := "test-sliding-window-clear"
	capacity := 2
	windowSize := 500 * time.Millisecond

	storage.SetDefaultValue(identifier, internal.SlidingWindowLogCtrParams{Capacity: capacity, WindowSize: windowSize})
	handler := internal.NewHandler(storage)

	// First request
	response, err := handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error on first request: %v", err)
	}
	if !response.IsAllowed() {
		t.Error("First request should be allowed")
	}

	time.Sleep(windowSize / 2)

	// Second request
	response, err = handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error on second request: %v", err)
	}
	if !response.IsAllowed() {
		t.Error("Second request should be allowed")
	}

	// Third request should be denied
	response, err = handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error on third request: %v", err)
	}
	if response.IsAllowed() {
		t.Error("Third request should be denied (capacity exceeded)")
	}

	// Wait for the first request to exit the window
	time.Sleep(windowSize/2 + 50*time.Millisecond) // total sleep is now windowSize/2 + windowSize/2 + 50ms = windowSize + 50ms

	// Now a new request should be allowed as the first one expired
	response, err = handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error after first request expired: %v", err)
	}
	if !response.IsAllowed() {
		t.Error("Request should be allowed after first request expired")
	}

	// Again, should be denied
	response, err = handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error on denied request: %v", err)
	}
	if response.IsAllowed() {
		t.Error("Request should be denied (capacity exceeded again)")
	}
}

func TestSlidingWindowLogAlgorithm_EmptyLogs(t *testing.T) {
	storage := internal.NewSlidingWindowLogMemoryStorage()
	identifier := "test-sliding-window-empty-logs"
	capacity := 1
	windowSize := 1 * time.Second

	storage.SetDefaultValue(identifier, internal.SlidingWindowLogCtrParams{Capacity: capacity, WindowSize: windowSize})
	handler := internal.NewHandler(storage)

	// Make one request
	response, err := handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !response.IsAllowed() {
		t.Error("Expected first request to be allowed")
	}

	// Wait for window to clear
	time.Sleep(windowSize + 50*time.Millisecond)

	// Make another request, should reinitialize logs
	response, err = handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error after window clear: %v", err)
	}
	if !response.IsAllowed() {
		t.Error("Expected request to be allowed after window clear and logs reinit")
	}
}
