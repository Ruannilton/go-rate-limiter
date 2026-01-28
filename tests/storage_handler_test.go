package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

// Mock storage to test handler error conditions
type MockErrorStorage[T internal.Algorithm] struct{}

func (m *MockErrorStorage[T]) SetDefaultValue(key string, value any) {}
func (m *MockErrorStorage[T]) New(key string) (T, error) {
	var zero T
	return zero, errors.New("storage.New failed")
}
func (m *MockErrorStorage[T]) Store(key string, value T) error { return nil }
func (m *MockErrorStorage[T]) Retrieve(key string) (T, bool) {
	var zero T
	return zero, false
}
func (m *MockErrorStorage[T]) Delete(key string) error { return nil }

func TestHandlerWrapper(t *testing.T) {
	storage := internal.NewFixedWindowMemoryStorage()
	identifier := "handler-test-isolated"
	capacity := 3
	storage.SetDefaultValue(identifier, internal.FixedWindowCtrParams{Capacity: capacity})
	handler := internal.NewHandler(storage)

	// Consume capacity - 1
	for i := 0; i < capacity-1; i++ {
		res, err := handler.Handle(identifier)
		if err != nil {
			t.Fatalf("handle call #%d failed: %v", i+1, err)
		}
		if !res.IsAllowed() {
			t.Fatalf("call #%d should have been allowed", i+1)
		}
	}

	// Last allowed call
	res, err := handler.Handle(identifier)
	if err != nil {
		t.Fatalf("last allowed handle call failed: %v", err)
	}
	if !res.IsAllowed() {
		t.Error("expected last call within capacity to be allowed")
	}

	// First denied call
	res, err = handler.Handle(identifier)
	if err != nil {
		t.Fatalf("first denied handle call failed: %v", err)
	}
	if res.IsAllowed() {
		t.Error("expected call beyond capacity to be denied")
	}
}

func TestHandlerWrapper_NewError(t *testing.T) {
	// Test that if storage.New() fails, the error is propagated.
	// We can trigger this by not setting a default value.
	storage := internal.NewFixedWindowMemoryStorage()
	identifier := "handler-new-error"
	handler := internal.NewHandler(storage)

	_, err := handler.Handle(identifier)
	if err == nil {
		t.Fatal("expected an error when New() fails, but got none")
	}
}

func TestFixedWindowMemoryStorage(t *testing.T) {
	s1 := internal.NewFixedWindowMemoryStorage()
	s2 := internal.NewFixedWindowMemoryStorage()
	if s1 != s2 {
		t.Error("NewFixedWindowMemoryStorage should return a singleton")
	}

	storage := internal.NewFixedWindowMemoryStorage()
	identifier := "fw-storage-test"
	params := internal.FixedWindowCtrParams{Capacity: 10}

	storage.SetDefaultValue(identifier, params)
	instance, err := storage.New(identifier)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if instance == nil {
		t.Fatal("New() returned a nil instance")
	}
	// Check if the capacity is correct by evaluating it
	for i := 0; i < params.Capacity; i++ {
		response := instance.Eval()
		if !(&response).IsAllowed() {
			t.Errorf("instance from New() should have capacity %d, failed at %d", params.Capacity, i)
			return
		}
	}
	response := instance.Eval()
	if (&response).IsAllowed() {
		t.Errorf("instance from New() has incorrect capacity, should have denied")
	}

	_, err = storage.New("no-default-key")
	if err == nil {
		t.Error("New() should fail without a default value")
	}

	manualInstance := internal.NewFixedWindowAlgorithmData("manual", 5)
	storage.Store("manual", manualInstance)
	retrieved, exists := storage.Retrieve("manual")
	if !exists {
		t.Fatal("Retrieve() failed to find stored item")
	}
	if retrieved != manualInstance {
		t.Error("Retrieve() returned wrong instance")
	}

	storage.Delete("manual")
	_, exists = storage.Retrieve("manual")
	if exists {
		t.Error("Delete() did not remove the item")
	}
}

func TestLeakyBucketMemoryStorage(t *testing.T) {
	s1 := internal.NewLeakyBucketMemoryStorage()
	s2 := internal.NewLeakyBucketMemoryStorage()
	if s1 != s2 {
		t.Error("NewLeakyBucketMemoryStorage should return a singleton")
	}

	storage := internal.NewLeakyBucketMemoryStorage()
	identifier := "lb-storage-test"
	params := internal.LeakyBucketCtrParams{Capacity: 10, DropPerSecond: 100}

	storage.SetDefaultValue(identifier, params)
	instance, err := storage.New(identifier)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if instance == nil {
		t.Fatal("New() returned a nil instance")
	}

	_, err = storage.New("no-default-key")
	if err == nil {
		t.Error("New() should fail without a default value")
	}
}

func TestTokenBucketMemoryStorage(t *testing.T) {
	s1 := internal.NewTokenBucketMemoryStorage()
	s2 := internal.NewTokenBucketMemoryStorage()
	if s1 != s2 {
		t.Error("NewTokenBucketMemoryStorage should return a singleton")
	}

	storage := internal.NewTokenBucketMemoryStorage()
	identifier := "tb-storage-test"
	params := internal.TokenBucketCtrParams{Capacity: 10, RefillRate: 10, RequestCost: 1}

	storage.SetDefaultValue(identifier, params)
	instance, err := storage.New(identifier)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if instance == nil {
		t.Fatal("New() returned a nil instance")
	}

	_, err = storage.New("no-default-key")
	if err == nil {
		t.Error("New() should fail without a default value")
	}
}

func TestSlidingWindowLogMemoryStorage(t *testing.T) {
	s1 := internal.NewSlidingWindowLogMemoryStorage()
	s2 := internal.NewSlidingWindowLogMemoryStorage()
	if s1 != s2 {
		t.Error("NewSlidingWindowLogMemoryStorage should return a singleton")
	}

	storage := internal.NewSlidingWindowLogMemoryStorage()
	identifier := "swl-storage-test"
	params := internal.SlidingWindowLogCtrParams{Capacity: 5, WindowSize: 1 * time.Second}

	storage.SetDefaultValue(identifier, params)
	instance, err := storage.New(identifier)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if instance == nil {
		t.Fatal("New() returned a nil instance")
	}

	for i := 0; i < params.Capacity; i++ {
		response := instance.Eval()
		if !(&response).IsAllowed() {
			t.Errorf("instance from New() should have capacity %d, failed at %d", params.Capacity, i)
			return
		}
	}
	response := instance.Eval()
	if (&response).IsAllowed() {
		t.Errorf("instance from New() has incorrect capacity, should have denied")
	}

	_, err = storage.New("no-default-key")
	if err == nil {
		t.Error("New() should fail without a default value")
	}

	manualInstance := internal.NewSlidingWindowLogAlgorithmData(1, 1*time.Minute)
	storage.Store("manual-swl", manualInstance)

	retrieved, exists := storage.Retrieve("manual-swl")
	if !exists {
		t.Fatal("Retrieve() failed to find stored item")
	}
	if retrieved != manualInstance {
		t.Error("Retrieve() returned wrong instance")
	}

	storage.Delete("manual-swl")
	_, exists = storage.Retrieve("manual-swl")
	if exists {
		t.Error("Delete() did not remove the item")
	}
}
