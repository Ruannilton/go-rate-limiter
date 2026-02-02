package rate_limiter

import (
	"errors"
	"sync"
	"time"
)

type slidingWindowLogLimiter struct {
	logs       []int64
	capacity   int
	windowSize time.Duration
	mutex      sync.Mutex
}

func newSlidingWindowLogLimiter(capacity int, windowSize time.Duration) *slidingWindowLogLimiter {
	return &slidingWindowLogLimiter{
		logs:       make([]int64, capacity),
		capacity:   capacity,
		windowSize: windowSize,
		mutex:      sync.Mutex{},
	}
}

func (s *slidingWindowLogLimiter) eval() RequestPipelineResponse {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-s.windowSize)

	newLogs := make([]int64, 0, s.capacity)
	for _, timestamp := range s.logs {
		if timestamp >= windowStart.UnixNano() {
			newLogs = append(newLogs, timestamp)
		}
	}
	s.logs = newLogs

	if len(s.logs) >= s.capacity {
		return newSyncRequestPipelineResponse(false)
	}

	// Add new request timestamp
	s.logs = append(s.logs, now.UnixNano())
	return newSyncRequestPipelineResponse(true)
}

type slidingWindowLogLimiterParams struct {
	Capacity   int
	WindowSize time.Duration
}

func getSlidingWindowLogLimiterParamsFromMap(params map[string]any) (slidingWindowLogLimiterParams, error) {
	capacity, capacityOk := getNumberFromMap[int](params, "capacity")
	if !capacityOk {
		return slidingWindowLogLimiterParams{}, errors.New("invalid capacity parameter")
	}
	windowSizeSeconds, intervalOk := getNumberFromMap[float64](params, "window_size")
	if !intervalOk {
		return slidingWindowLogLimiterParams{}, errors.New("invalid window_size parameter")
	}
	windowSize := time.Duration(windowSizeSeconds * float64(time.Second))
	return slidingWindowLogLimiterParams{
		Capacity:   capacity,
		WindowSize: windowSize,
	}, nil
}
