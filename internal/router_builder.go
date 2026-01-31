package internal

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

type StrategyName string

const (
	LimiterStrategyFixedWindow      StrategyName = "fixed_window"
	LimiterStrategyTokenBucket      StrategyName = "token_bucket"
	LimiterStrategySlidingWindowLog StrategyName = "sliding_window_log"
	TrafficStrategyLeakyBucket      StrategyName = "leaky_bucket"
)

type StrategyDescriptor struct {
    StrategyName StrategyName   `json:"type" yaml:"type"`
    Params       map[string]any `json:"params" yaml:"params"`
}

type RouteDescriptor struct {
    Path                    string              `json:"path" yaml:"path"`
    LimiterDescriptor       *StrategyDescriptor `json:"limiter,omitempty" yaml:"limiter,omitempty"`
    TrafficShaperDescriptor *StrategyDescriptor `json:"traffic,omitempty" yaml:"traffic,omitempty"`
}

type RouterBuilder struct {
	router *Router
}

func NewRouterBuilder() *RouterBuilder {
	return &RouterBuilder{
		router: NewRouter(),
	}
}

func (r *Router) AddRoute(route RouteDescriptor, closeSign <-chan struct{}) error {
	var lim IRateLimiter
	var traf ITrafficShapeAlgorithm

	if route.LimiterDescriptor != nil {
		limiter, err := createRateLimiterFromDescriptor(*route.LimiterDescriptor)
		if err != nil {
			return err
		}
		lim = limiter
	}

	if route.TrafficShaperDescriptor != nil {
		shapper, err := createTrafficShaperFromDescriptor(*route.TrafficShaperDescriptor, closeSign)
		if err != nil {
			return err
		}
		traf = shapper
	}

	pipeline := NewRequestPipeline(lim, traf)
	r.setupPath(route.Path, pipeline)
	return nil
}

func (r *Router) LoadFromJson(jsonData []byte, closeSign <-chan struct{}) error {
	
    var descriptors []RouteDescriptor

    if err := json.Unmarshal(jsonData, &descriptors); err != nil {
        return fmt.Errorf("falha ao ler JSON: %w", err)
    }

    for _, routeDesc := range descriptors {
        if err := r.AddRoute(routeDesc, closeSign); err != nil {
            return fmt.Errorf("erro ao adicionar rota '%s': %w", routeDesc.Path, err)
        }
    }

    return nil
}

func (r *Router) LoadFromYaml(yamlData []byte, closeSign <-chan struct{}) error {

    var descriptors []RouteDescriptor

    if err := yaml.Unmarshal(yamlData, &descriptors); err != nil {
        return fmt.Errorf("falha ao ler YAML: %w", err)
    }

    for _, routeDesc := range descriptors {
        if err := r.AddRoute(routeDesc, closeSign); err != nil {
            return fmt.Errorf("erro ao adicionar rota '%s': %w", routeDesc.Path, err)
        }
    }

    return nil
}

func createTrafficShaperFromDescriptor(strategyDescriptor StrategyDescriptor, closeSign <-chan struct{}) (ITrafficShapeAlgorithm, error) {
	switch strategyDescriptor.StrategyName {
	case TrafficStrategyLeakyBucket:
		params, err := GetLeakyBucketTrafficShaperParamsFromMap(strategyDescriptor.Params)
		if err != nil {
			return nil, err
		}
		return NewLeakyBucketTrafficShaper(params.Capacity, params.DropPerSecond, closeSign), nil
	default:
		return nil, errors.New("unknown traffic shaper strategy")
	}
}

func createRateLimiterFromDescriptor(routeLimiterDescriptor StrategyDescriptor) (IRateLimiter, error) {

	switch routeLimiterDescriptor.StrategyName {
	case LimiterStrategyFixedWindow:
		params, err := GetFixedWindowRateLimiterParamsFromMap(routeLimiterDescriptor.Params)
		if err != nil {
			return nil, err
		}
		return NewFixedWindowRateLimiter(params.Capacity, params.ResetInterval), nil
	case LimiterStrategyTokenBucket:
		params, err := GetTokenBucketRateLimiterParamsFromMap(routeLimiterDescriptor.Params)
		if err != nil {
			return nil, err
		}
		return NewTokenBucketRateLimiter(params.Capacity, params.RefillRate, params.RequestCost), nil
	case LimiterStrategySlidingWindowLog:
		params, err := GetSlidingWindowLogLimiterParamsFromMap(routeLimiterDescriptor.Params)
		if err != nil {
			return nil, err
		}
		return NewSlidingWindowLogLimiter(params.Capacity, params.WindowSize), nil
	default:
		return nil, errors.New("unknown limiter strategy")
	}

}
