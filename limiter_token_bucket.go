package rate_limiter

import (
	"errors"
	"sync"
	"time"
)

type tokenBucketRateLimiter struct {
	capacity          float64
	tokens            float64
	requestCost       float64
	refillRateSeconds float64
	mutex             sync.Mutex
	lastRefill        time.Time
}

func newTokenBucketRateLimiter(capacity, refillRate, requestCost float64) *tokenBucketRateLimiter {
	return &tokenBucketRateLimiter{
		capacity:          capacity,
		tokens:            capacity,
		refillRateSeconds: refillRate,
		mutex:             sync.Mutex{},
		lastRefill:        time.Now(),
		requestCost:       requestCost,
	}
}

func (t *tokenBucketRateLimiter) eval() RequestPipelineResponse {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tokensToAdd := float64(time.Since(t.lastRefill).Milliseconds()) * t.refillRateSeconds / 1000
	t.lastRefill = time.Now()
	t.tokens = min(t.capacity, t.tokens+tokensToAdd)
	if t.tokens >= t.requestCost {
		t.tokens -= t.requestCost
		return newSyncRequestPipelineResponse(true)
	} else {
		return newSyncRequestPipelineResponse(false)
	}
}

type tokenBucketRateLimiterParams struct {
	Capacity    float64
	RefillRate  float64
	RequestCost float64
}

func getTokenBucketRateLimiterParamsFromMap(params map[string]any) (tokenBucketRateLimiterParams, error) {
	capacity, capacityOk := getNumberFromMap[float64](params, "capacity")
	if !capacityOk {
		return tokenBucketRateLimiterParams{}, errors.New("invalid capacity parameter")
	}
	refillRate, refillOk := getNumberFromMap[float64](params, "refill_rate")
	if !refillOk {
		return tokenBucketRateLimiterParams{}, errors.New("invalid refill_rate parameter")
	}
	requestCost, requestCostOk := getNumberFromMap[float64](params, "request_cost")
	if !requestCostOk {
		return tokenBucketRateLimiterParams{}, errors.New("invalid request_cost parameter")
	}
	return tokenBucketRateLimiterParams{
		Capacity:    capacity,
		RefillRate:  refillRate,
		RequestCost: requestCost,
	}, nil
}
