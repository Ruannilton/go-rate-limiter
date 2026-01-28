package tests

import (
	"testing"
	"time"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func TestTokenBucketAlgorithm(t *testing.T) {
	storage := internal.NewTokenBucketMemoryStorage()
	identifier := "test-token-bucket"
	capacity := 10.0
	refillRate := 10.0 // 10 tokens per second
	requestCost := 1.0
	storage.SetDefaultValue(identifier, internal.TokenBucketCtrParams{Capacity: capacity, RefillRate: refillRate, RequestCost: requestCost})
	handler := internal.NewHandler(storage)

	// Initial state has full capacity
	for i := 0; i < int(capacity); i++ {
		response, err := handler.Handle(identifier)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !response.IsAllowed() {
			t.Errorf("Request %d should have been allowed", i+1)
		}
	}

	// Next request should be denied
	response, err := handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if response.IsAllowed() {
		t.Error("Request should have been denied")
	}

	// Wait for refill
	time.Sleep(100 * time.Millisecond)

	response, err = handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !response.IsAllowed() {
		t.Error("Request should have been allowed after refill")
	}
}

func TestTokenBucketAlgorithmRefill(t *testing.T) {
	storage := internal.NewTokenBucketMemoryStorage()
	identifier := "test-token-bucket-refill"
	capacity := 5.0
	refillRate := 5.0 // 5 tokens per second
	requestCost := 1.0
	storage.SetDefaultValue(identifier, internal.TokenBucketCtrParams{Capacity: capacity, RefillRate: refillRate, RequestCost: requestCost})
	handler := internal.NewHandler(storage)

	// Consume all tokens
	for i := 0; i < int(capacity); i++ {
		_, err := handler.Handle(identifier)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	// Should be empty now
	response, err := handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if response.IsAllowed() {
		t.Error("Bucket should be empty")
	}

	// Wait for 2 tokens to be refilled
	time.Sleep(400 * time.Millisecond)

	// Should allow 2 requests
	for i := 0; i < 2; i++ {
		response, err := handler.Handle(identifier)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !response.IsAllowed() {
			t.Errorf("Request %d after refill should be allowed", i+1)
		}
	}

	// Next one should be denied
	response, err = handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if response.IsAllowed() {
		t.Error("Request should be denied after consuming refilled tokens")
	}
}
