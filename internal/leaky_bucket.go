package internal

import (
	"sync"
	"time"
)

type LeakyBucketCtrParams struct {
	Capacity      int
	DropPerSecond int64
}

var defaultLeakyBucketParam = LeakyBucketCtrParams{
	Capacity:      200,
	DropPerSecond: 100,
}


type LeakyBucketAlgorithmData struct {
	identifier           string
	requestShapingTicker *time.Ticker
	queue                chan chan bool
	lastTimeAdded        time.Time
	mutex                sync.Mutex
}

const LEAKY_BUCKET_EXPIRATION_CHECK_INTERVAL = time.Minute * 1

func NewLeakyBucketAlgorithmData(identifier string, capacity int, dropPerSecond int64) *LeakyBucketAlgorithmData {
	intervalo := time.Second / time.Duration(dropPerSecond)

	algo := &LeakyBucketAlgorithmData{
		identifier:           identifier,
		queue:                make(chan chan bool, capacity),
		requestShapingTicker: time.NewTicker(intervalo),
		mutex:                sync.Mutex{},
		lastTimeAdded:        time.Now(),
	}

	go func(bucket *LeakyBucketAlgorithmData) {
		expirationCheck := time.NewTicker(LEAKY_BUCKET_EXPIRATION_CHECK_INTERVAL)
		defer expirationCheck.Stop()
		for {
			select {
			case <-bucket.requestShapingTicker.C:
				select {
				case responseChannel := <-bucket.queue:
					responseChannel <- true
				default:
				}

			case <-expirationCheck.C:
				bucket.mutex.Lock()
				expirated :=  (bucket.lastTimeAdded.IsZero() || time.Since(bucket.lastTimeAdded) > EXPIRATION_CHECK_INTERVAL)
				empty:= len(bucket.queue) == 0
				bucket.mutex.Unlock()
				if expirated && empty {
					bucket.requestShapingTicker.Stop()
					return
				}
			}
		}
	}(algo)
	return algo
}

func (bucket *LeakyBucketAlgorithmData) Eval() AlgorithmResponse {
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()
	responseChannel := make(chan bool, 1)
	bucket.lastTimeAdded = time.Now()
	select {
	case bucket.queue <- responseChannel:
		return AlgorithmResponse{
			Allowed:             true,
			Delayed:             true,
			DelayedResponseChan: responseChannel,
		}
	default:
		return AlgorithmResponse{
			Allowed:             false,
			Delayed:             false,
			DelayedResponseChan: nil,
		}
	}
}

type LeakyBucketMemoryStorage struct {
	currentBuckets map[string]*LeakyBucketAlgorithmData
	oldBuckets     map[string]*LeakyBucketAlgorithmData
	defaults       map[string]any
	mutexLock      sync.Mutex
}

var leakyBucketMemoryStorageSingleton *LeakyBucketMemoryStorage
var once sync.Once

func NewLeakyBucketMemoryStorage() *LeakyBucketMemoryStorage {
	once.Do(func() {
		leakyBucketMemoryStorageSingleton = newLeakyBucketMemoryStorage()
	})
	return leakyBucketMemoryStorageSingleton
}

func newLeakyBucketMemoryStorage() *LeakyBucketMemoryStorage {

	storage := &LeakyBucketMemoryStorage{
		currentBuckets: make(map[string]*LeakyBucketAlgorithmData),
		oldBuckets:     make(map[string]*LeakyBucketAlgorithmData),
		defaults:       make(map[string]any),
		mutexLock:      sync.Mutex{},
	}
	storage.startCleanupRoutine()
	return storage
}

func (storage *LeakyBucketMemoryStorage) SetDefaultValue(key string, value any) {
	storage.mutexLock.Lock()
	defer storage.mutexLock.Unlock()
	storage.defaults[key] = value
}

func (storage *LeakyBucketMemoryStorage) New(key string) (*LeakyBucketAlgorithmData, error) {
	params := storage.getConstructorParams(key)
	return NewLeakyBucketAlgorithmData(key, params.Capacity, int64(params.DropPerSecond)), nil
}

func (storage *LeakyBucketMemoryStorage) Store(key string, value *LeakyBucketAlgorithmData) error {
	storage.mutexLock.Lock()
	defer storage.mutexLock.Unlock()
	storage.currentBuckets[key] = value
	return nil
}

func (storage *LeakyBucketMemoryStorage) Retrieve(key string) (*LeakyBucketAlgorithmData, bool) {
	storage.mutexLock.Lock()
	defer storage.mutexLock.Unlock()
	val, exists := storage.currentBuckets[key]
	if !exists {
		val, exists = storage.oldBuckets[key]
		if exists {
			storage.currentBuckets[key] = val
			delete(storage.oldBuckets, key)
		}
	}
	return val, exists
}

func (storage *LeakyBucketMemoryStorage) Delete(key string) error {
	storage.mutexLock.Lock()
	defer storage.mutexLock.Unlock()
	delete(storage.currentBuckets, key)
	delete(storage.oldBuckets, key)
	return nil
}

func (storage *LeakyBucketMemoryStorage) startCleanupRoutine() {
	go func() {
		timer := time.NewTicker(EXPIRATION_CHECK_INTERVAL)

		for {
			<-timer.C
			storage.mutexLock.Lock()
			storage.oldBuckets = storage.currentBuckets
			storage.currentBuckets = make(map[string]*LeakyBucketAlgorithmData)
			storage.mutexLock.Unlock()
		}
	}()
}

func (storage *LeakyBucketMemoryStorage) getConstructorParams(key string) LeakyBucketCtrParams {
	storage.mutexLock.Lock()
	defer storage.mutexLock.Unlock()
	if def, exists := storage.defaults[key]; exists {
		if params, ok := def.(LeakyBucketCtrParams); ok {
			return params
		}
	}
	return LeakyBucketCtrParams{
		Capacity: 200,
		DropPerSecond: 100,
	}
}