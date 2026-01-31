package internal

import (
	"errors"
	"sync"
	"time"
)

type SlidingWindowLogLimiter struct {
	logs       []int64
	capacity   int
	windowSize time.Duration
	mutex      sync.Mutex
}

func NewSlidingWindowLogLimiter(capacity int, windowSize time.Duration) *SlidingWindowLogLimiter {
	return &SlidingWindowLogLimiter{
		logs:       make([]int64, capacity),
		capacity:   capacity,
		windowSize: windowSize,
		mutex:      sync.Mutex{},
	}
}


func (s *SlidingWindowLogLimiter) Eval() RequestPipelineResponse {
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
		return NewSyncRequestPipelineResponse(false)
	}

	// Add new request timestamp
	s.logs = append(s.logs, now.UnixNano())
	return NewSyncRequestPipelineResponse(true)
}

type SlidingWindowLogLimiterParams struct {
	Capacity   int
	WindowSize time.Duration
}

func GetSlidingWindowLogLimiterParamsFromMap(params map[string]any) (SlidingWindowLogLimiterParams,error) {
	capacity, capacityOk := getNumberFromMap[int](params, "capacity")
	if !capacityOk {
		return SlidingWindowLogLimiterParams{},errors.New("invalid capacity parameter")
	}
	windowSizeSeconds, intervalOk := getNumberFromMap[float64](params, "window_size")
	if !intervalOk {
		return SlidingWindowLogLimiterParams{},errors.New("invalid window_size parameter")
	}
	windowSize := time.Duration(windowSizeSeconds * float64(time.Second))
	return SlidingWindowLogLimiterParams{
		Capacity:   capacity,
		WindowSize: windowSize,
	},nil
}