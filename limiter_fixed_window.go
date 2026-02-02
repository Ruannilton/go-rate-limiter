package rate_limiter

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type fixedWindowRateLimiter struct {
	counter       int
	capacity      int
	mutex         sync.Mutex
	lastReset     time.Time
	resetInterval time.Duration
}

func newFixedWindowRateLimiter(capacity int, resetInterval time.Duration) *fixedWindowRateLimiter {
	return &fixedWindowRateLimiter{
		counter:       0,
		capacity:      capacity,
		mutex:         sync.Mutex{},
		lastReset:     time.Now(),
		resetInterval: resetInterval,
	}
}

func (f *fixedWindowRateLimiter) eval() RequestPipelineResponse {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if time.Since(f.lastReset) >= f.resetInterval {
		f.counter = 0
		f.lastReset = time.Now()
	}
	if f.counter < f.capacity {
		f.counter++
		return newSyncRequestPipelineResponse(true)
	} else {
		return newSyncRequestPipelineResponse(false)
	}
}

type FixedWindowRateLimiterParams struct {
	Capacity      int
	ResetInterval time.Duration
}

func GetFixedWindowRateLimiterParamsFromMap(params map[string]any) (FixedWindowRateLimiterParams, error) {
	capacity, capacityOk := getNumberFromMap[int](params, "capacity")
	if !capacityOk {
		msg := fmt.Sprintf("invalid capacity parameter: %v", params["capacity"])
		return FixedWindowRateLimiterParams{}, errors.New(msg)
	}
	resetIntervalSeconds, intervalOk := getNumberFromMap[float64](params, "reset_interval")
	if !intervalOk {
		msg := fmt.Sprintf("invalid reset_interval parameter: %v", params["reset_interval"])
		return FixedWindowRateLimiterParams{}, errors.New(msg)
	}
	resetInterval := time.Duration(resetIntervalSeconds * float64(time.Second))
	return FixedWindowRateLimiterParams{
		Capacity:      capacity,
		ResetInterval: resetInterval,
	}, nil
}
