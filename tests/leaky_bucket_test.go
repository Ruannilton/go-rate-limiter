package tests

import (
	"testing"
	"time"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func TestLeakyBucketAlgorithm(t *testing.T) {
	storage := internal.NewLeakyBucketMemoryStorage()
	identifier := "test-leaky-bucket"
	capacity := 10
	dropPerSecond := int64(10)
	storage.SetDefaultValue(identifier, internal.LeakyBucketCtrParams{Capacity: capacity, DropPerSecond: dropPerSecond})
	handler := internal.NewHandler(storage)

	for i := 0; i < capacity; i++ {
		response, err := handler.Handle(identifier)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !response.Allowed {
			t.Errorf("Request %d should have been allowed, but was not", i+1)
		}
		if !response.Delayed {
			t.Errorf("Request %d should have been delayed, but was not", i+1)
		}
	}

	response, err := handler.Handle(identifier)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if response.Allowed {
		t.Error("Request should have been denied, but was allowed")
	}
}

func TestLeakyBucketAlgorithmProcessing(t *testing.T) {
	storage := internal.NewLeakyBucketMemoryStorage()
	identifier := "test-leaky-bucket-processing"
	capacity := 5
	dropPerSecond := int64(10) // 1 request every 100ms
	storage.SetDefaultValue(identifier, internal.LeakyBucketCtrParams{Capacity: capacity, DropPerSecond: dropPerSecond})
	handler := internal.NewHandler(storage)

	var responses []internal.AlgorithmResponse

	for i := 0; i < capacity; i++ {
		response, err := handler.Handle(identifier)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		responses = append(responses, response)
	}

	for i, resp := range responses {
		if !resp.Delayed {
			t.Errorf("Request %d should be delayed", i)
			continue
		}

		select {
		case allowed := <-resp.DelayedResponseChan:
			if !allowed {
				t.Errorf("Request %d should have been allowed", i)
			}
		case <-time.After(200 * time.Millisecond):
			t.Errorf("Request %d timed out", i)
		}
	}
}
