package internal

import (
	"sync"
	"time"
)

type slidingWindowLog []int64

type SlidingWindowLogAlgorithmData struct {
	logs       slidingWindowLog
	capacity   int
	windowSize time.Duration
	mutex      sync.Mutex
}

type SlidingWindowLogCtrParams struct {
	Capacity int
	WindowSize time.Duration
}

func NewSlidingWindowLogAlgorithmData(capacity int, windowSize time.Duration) *SlidingWindowLogAlgorithmData {
	return &SlidingWindowLogAlgorithmData{
		logs:       make(slidingWindowLog, 0),
		capacity:   capacity,
		windowSize: windowSize,
		mutex:      sync.Mutex{},
	}
}


func (s *SlidingWindowLogAlgorithmData) Eval() AlgorithmResponse {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-s.windowSize)

	// Clean up old log entries
	newLogs := make(slidingWindowLog, 0, s.capacity)
	for _, timestamp := range s.logs {
		if timestamp >= windowStart.UnixNano() {
			newLogs = append(newLogs, timestamp)
		}
	}
	s.logs = newLogs

	// Check capacity
	if len(s.logs) >= s.capacity {
		return AlgorithmResponse{
			Allowed: false,
		}
	}

	// Add new request timestamp
	s.logs = append(s.logs, now.UnixNano())
	return AlgorithmResponse{
		Allowed: true,
	}
}

type SlidingWindowLogMemoryStorage struct {
	currentBuckets map[string]*SlidingWindowLogAlgorithmData
	oldBuckets     map[string]*SlidingWindowLogAlgorithmData
	defaults       map[string]any
	mutexLock      sync.Mutex
}

var (
	slidingWindowLogMemoryStorageSingleton *SlidingWindowLogMemoryStorage
	slidingWindowOnce                      sync.Once
)

func NewSlidingWindowLogMemoryStorage() *SlidingWindowLogMemoryStorage {
	slidingWindowOnce.Do(func() {
		slidingWindowLogMemoryStorageSingleton = newSlidingWindowLogMemoryStorage()
	})
	return slidingWindowLogMemoryStorageSingleton
}
func newSlidingWindowLogMemoryStorage() *SlidingWindowLogMemoryStorage {
	storage := &SlidingWindowLogMemoryStorage{
		currentBuckets: make(map[string]*SlidingWindowLogAlgorithmData),
		oldBuckets:     make(map[string]*SlidingWindowLogAlgorithmData),
		defaults:       make(map[string]any),
		mutexLock:      sync.Mutex{},
	}
	storage.startCleanupRoutine()
	return storage
}

func (self *SlidingWindowLogMemoryStorage) SetDefaultValue(key string, value any) {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	self.defaults[key] = value
}

func (self *SlidingWindowLogMemoryStorage) New(key string) (*SlidingWindowLogAlgorithmData, error) {
	params := self.getConstructorParams(key)
	return NewSlidingWindowLogAlgorithmData(params.Capacity, params.WindowSize), nil
}

func (self *SlidingWindowLogMemoryStorage) Store(key string, value *SlidingWindowLogAlgorithmData) error {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	self.currentBuckets[key] = value
	return nil
}
func (self *SlidingWindowLogMemoryStorage) Retrieve(key string) (*SlidingWindowLogAlgorithmData, bool) {
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

func (self *SlidingWindowLogMemoryStorage) Delete(key string) error {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	delete(self.currentBuckets, key)
	delete(self.oldBuckets, key)
	return nil
}

func (self *SlidingWindowLogMemoryStorage) startCleanupRoutine() {
	go func() {
		timer := time.NewTicker(EXPIRATION_CHECK_INTERVAL)
		for {
			<-timer.C
			self.mutexLock.Lock()
			self.oldBuckets = self.currentBuckets
			self.currentBuckets = make(map[string]*SlidingWindowLogAlgorithmData)
			self.mutexLock.Unlock()
		}
	}()
}

func (self *SlidingWindowLogMemoryStorage) getConstructorParams(key string) SlidingWindowLogCtrParams {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	if def, exists := self.defaults[key]; exists {
		if params, ok := def.(SlidingWindowLogCtrParams); ok {
			return params
		}
	}
	return SlidingWindowLogCtrParams{
		Capacity:   100,
		WindowSize: time.Minute,
	}
}