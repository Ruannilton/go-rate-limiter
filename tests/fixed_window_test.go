package tests

import (
	"testing"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func TestFixedWindowAlgorithm(t *testing.T) {
	storage := internal.NewFixedWindowMemoryStorage()
	identifier := "test-fixed-window"
	capacity := 5
	storage.SetDefaultValue(identifier, internal.FixedWindowCtrParams{Capacity: capacity})
	handler := internal.NewHandler(storage)

	// Allow requests up to capacity
	for i := 0; i < capacity; i++ {
		response, err := handler.Handle(identifier)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !response.IsAllowed() {
			t.Errorf("Request %d should have been allowed, but was denied", i+1)
		}
	}

	// Deny next request
	response, err := handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if response.IsAllowed() {
		t.Error("Request should have been denied, but was allowed")
	}
}

func TestFixedWindowAlgorithmReset(t *testing.T) {
	storage := internal.NewFixedWindowMemoryStorage()
	identifier := "test-fixed-window-reset"
	capacity := 5
	storage.SetDefaultValue(identifier, internal.FixedWindowCtrParams{Capacity: capacity})
	handler := internal.NewHandler(storage)

	for i := 0; i < capacity; i++ {
		_, err := handler.Handle(identifier)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	response, err := handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if response.IsAllowed() {
		t.Error("Request should have been denied after filling capacity")
	}

	// Delete and try again
	storage.Delete(identifier)

	// Now it should allow again
	response, err = handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !response.IsAllowed() {
		t.Error("Request should have been allowed after reset")
	}
}
