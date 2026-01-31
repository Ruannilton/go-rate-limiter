package internal

import (
	"errors"
	"sync"
	"time"
)

type TokenBucketRateLimiter struct {
	capacity   float64
	tokens     float64
	requestCost float64
	refillRateSeconds float64
	mutex      sync.Mutex
	lastRefill time.Time
}

func NewTokenBucketRateLimiter(capacity, refillRate, requestCost float64) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		capacity:   capacity,
		tokens:     capacity,
		refillRateSeconds: refillRate,
		mutex:      sync.Mutex{},
		lastRefill: time.Now(),
		requestCost: requestCost,
	}
}

func (t *TokenBucketRateLimiter) Eval() RequestPipelineResponse {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tokensToAdd := float64(time.Since(t.lastRefill).Milliseconds()) * t.refillRateSeconds / 1000
	t.lastRefill = time.Now()
	t.tokens = min(t.capacity, t.tokens+tokensToAdd)
	if t.tokens >= t.requestCost {
		t.tokens -= t.requestCost
		return NewSyncRequestPipelineResponse(true)
	} else {
		return NewSyncRequestPipelineResponse(false)
	}
}

type TokenBucketRateLimiterParams struct {
	Capacity   float64
	RefillRate float64
	RequestCost float64
}

func GetTokenBucketRateLimiterParamsFromMap(params map[string]any) (TokenBucketRateLimiterParams,error) {
	capacity, capacityOk := getNumberFromMap[float64](params, "capacity")
	if !capacityOk {
		return TokenBucketRateLimiterParams{},errors.New("invalid capacity parameter")
	}
	refillRate, refillOk := getNumberFromMap[float64](params, "refill_rate")
	if !refillOk {
		return TokenBucketRateLimiterParams{},errors.New("invalid refill_rate parameter")
	}
	requestCost, requestCostOk := getNumberFromMap[float64](params, "request_cost")
	if !requestCostOk {
		return TokenBucketRateLimiterParams{},errors.New("invalid request_cost parameter")
	}
	return TokenBucketRateLimiterParams{
		Capacity:   capacity,
		RefillRate: refillRate,
		RequestCost: requestCost,
	},nil
}


