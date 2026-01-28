package internal

import (
	"sync"
	"time"
)

type TokenBucketCtrParams struct {
	Capacity    float64
	RefillRate  float64
	RequestCost float64
}

type TokenBucketAlgorithmData struct {
	currentValue        float64
	capacity            float64
	refilRate           float64
	requestCost         float64
	lastRefillTimestamp int64
	mutex               sync.Mutex
}

func NewTokenBucketAlgorithmData(capacity, refilRate, requestCost float64) *TokenBucketAlgorithmData {
	return &TokenBucketAlgorithmData{
		currentValue:        capacity,
		capacity:            capacity,
		refilRate:           refilRate,
		requestCost:         requestCost,
		lastRefillTimestamp: time.Now().UnixNano(),
		mutex:               sync.Mutex{},
	}
}

func (self *TokenBucketAlgorithmData) Eval() AlgorithmResponse {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	nowTimestamp := time.Now().UnixNano()
	tokensToAdd := float64(nowTimestamp-self.lastRefillTimestamp) * self.refilRate / 1e9
	self.currentValue = min(self.capacity, self.currentValue+tokensToAdd)
	self.lastRefillTimestamp = nowTimestamp

	if self.currentValue >= self.requestCost {
		self.currentValue -= self.requestCost
		return AlgorithmResponse{
			Allowed:             true,
			Delayed:             false,
			DelayedResponseChan: nil,
		}
	}

	return AlgorithmResponse{
		Allowed:             false,
		Delayed:             false,
		DelayedResponseChan: nil,
	}
}

type TokenBucketMemoryStorage struct {
	currentBuckets map[string]*TokenBucketAlgorithmData
	oldBuckets     map[string]*TokenBucketAlgorithmData
	defaults       map[string]any
	mutexLock      sync.Mutex
}

var (
	tokenBucketMemoryStorageSingleton *TokenBucketMemoryStorage
	tokenBucketOnce                   sync.Once
)

func NewTokenBucketMemoryStorage() *TokenBucketMemoryStorage {
	tokenBucketOnce.Do(func() {
		tokenBucketMemoryStorageSingleton = newTokenBucketMemoryStorage()
	})
	return tokenBucketMemoryStorageSingleton
}

func newTokenBucketMemoryStorage() *TokenBucketMemoryStorage {
	storage := &TokenBucketMemoryStorage{
		currentBuckets: make(map[string]*TokenBucketAlgorithmData),
		oldBuckets:     make(map[string]*TokenBucketAlgorithmData),
		defaults:       make(map[string]any),
		mutexLock:      sync.Mutex{},
	}
	storage.startCleanupRoutine()
	return storage
}

func (self *TokenBucketMemoryStorage) SetDefaultValue(key string, value any) {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	self.defaults[key] = value
}

func (self *TokenBucketMemoryStorage) New(key string) (*TokenBucketAlgorithmData, error) {
	params := self.getConstructorParams(key)
	return NewTokenBucketAlgorithmData(params.Capacity, params.RefillRate, params.RequestCost), nil
}

func (self *TokenBucketMemoryStorage) Store(key string, value *TokenBucketAlgorithmData) error {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	self.currentBuckets[key] = value
	return nil
}

func (self *TokenBucketMemoryStorage) Retrieve(key string) (*TokenBucketAlgorithmData, bool) {
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

func (self *TokenBucketMemoryStorage) Delete(key string) error {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	delete(self.currentBuckets, key)
	delete(self.oldBuckets, key)
	return nil
}

func (self *TokenBucketMemoryStorage) startCleanupRoutine() {
	go func() {
		timer := time.NewTicker(EXPIRATION_CHECK_INTERVAL)
		for {
			<-timer.C
			self.mutexLock.Lock()
			self.oldBuckets = self.currentBuckets
			self.currentBuckets = make(map[string]*TokenBucketAlgorithmData)
			self.mutexLock.Unlock()
		}
	}()
}

func (self *TokenBucketMemoryStorage) getConstructorParams(key string) TokenBucketCtrParams {
	self.mutexLock.Lock()
	defer self.mutexLock.Unlock()
	if def, exists := self.defaults[key]; exists {
		if params, ok := def.(TokenBucketCtrParams); ok {
			return params
		}
	}
	return TokenBucketCtrParams{
		Capacity:	100,
		RefillRate: 5,
		RequestCost: 1,
	}
}