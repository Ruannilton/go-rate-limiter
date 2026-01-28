package internal

import (
	"sync"
	"time"
)



type FixedWindowCtrParams struct {
	Capacity int
}


type FixedWindowAlgorithmData struct {
	identifier string
	counter    int
	capacity   int
	mutex      sync.Mutex
}

func NewFixedWindowAlgorithmData(identifier string, capacity int) *FixedWindowAlgorithmData {
	return &FixedWindowAlgorithmData{
		identifier: identifier,
		counter:    0,
		capacity:   capacity,
		mutex:      sync.Mutex{},
	}
}

func (self *FixedWindowAlgorithmData) Eval() AlgorithmResponse {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if self.counter < self.capacity {
		self.counter++
		return AlgorithmResponse{
			Allowed: true,
		}
	} else {
		return AlgorithmResponse{
			Allowed: false,
		}
	}
}

type FixedWindowMemoryStorage struct {
	currentBuckets map[string]*FixedWindowAlgorithmData
	oldBuckets     map[string]*FixedWindowAlgorithmData
	defaults       map[string]any
	mutexLock      sync.Mutex
}

var (
	fixedWindowMemoryStorageSingleton *FixedWindowMemoryStorage
	fixedWindowOnce                   sync.Once
)

func NewFixedWindowMemoryStorage() *FixedWindowMemoryStorage {
	fixedWindowOnce.Do(func() {
		fixedWindowMemoryStorageSingleton = newFixedWindowMemoryStorage()
	})
	return fixedWindowMemoryStorageSingleton
}
func newFixedWindowMemoryStorage() *FixedWindowMemoryStorage {
	storage := &FixedWindowMemoryStorage{
		currentBuckets: make(map[string]*FixedWindowAlgorithmData),
		oldBuckets:     make(map[string]*FixedWindowAlgorithmData),
		defaults:       make(map[string]any),
		mutexLock:      sync.Mutex{},
	}
	storage.startCleanupRoutine()
	return storage
}

func (self *FixedWindowMemoryStorage) SetDefaultValue(key string, value any) {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	self.defaults[key] = value
}

func (self *FixedWindowMemoryStorage) New(key string) (*FixedWindowAlgorithmData, error) {
	params:= self.getConstructorParams(key)
	return NewFixedWindowAlgorithmData(key, params.Capacity), nil
}

func (self *FixedWindowMemoryStorage) Store(key string, value *FixedWindowAlgorithmData) error {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	self.currentBuckets[key] = value
	return nil
}
func (self *FixedWindowMemoryStorage) Retrieve(key string) (*FixedWindowAlgorithmData, bool) {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	val, exists := self.currentBuckets[key]
	if !exists {
		val, exists = self.oldBuckets[key]
		if exists {
			self.currentBuckets[key] = val
			delete(self.oldBuckets, key)
		}
	}
	return val, exists
}
func (self *FixedWindowMemoryStorage) Delete(key string) error {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	delete(self.currentBuckets, key)
	delete(self.oldBuckets, key)
	return nil
}

func (self *FixedWindowMemoryStorage) startCleanupRoutine() {
	go func() {
		timer := time.NewTicker(EXPIRATION_CHECK_INTERVAL)
		for {
			<-timer.C
			self.mutexLock.Lock()
			self.oldBuckets = self.currentBuckets
			self.currentBuckets = make(map[string]*FixedWindowAlgorithmData)
			self.mutexLock.Unlock()
		}
	}()
}

func (self *FixedWindowMemoryStorage) getConstructorParams(key string) FixedWindowCtrParams {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	if def, exists := self.defaults[key]; exists {
		if params, ok := def.(FixedWindowCtrParams); ok {
			return params
		}
	}
	return FixedWindowCtrParams{
		Capacity: 10,
	}
}