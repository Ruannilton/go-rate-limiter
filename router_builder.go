package rate_limiter

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
	descriptors  map[string]RouteDescriptor
	closeSignal <-chan struct{}
}

func NewRouterBuilder(closeSign <-chan struct{}) RouterBuilder {
	return RouterBuilder{
		descriptors: make(map[string]RouteDescriptor),
		closeSignal: closeSign,
	}
}

func (r *RouterBuilder) Build() Router {
	router := newRouter()
	for _, route := range r.descriptors {
		router.setupRoute(route, r.closeSignal)
	}
	return router
}

func (r *RouterBuilder) SetRoute(route RouteDescriptor) {
	r.descriptors[route.Path] = route
}

func (r *RouterBuilder) RemoveRoute(path string) {
	delete(r.descriptors, path)
}

func (r *RouterBuilder) GetRouteDescriptors() []RouteDescriptor {
	descriptors := make([]RouteDescriptor, 0, len(r.descriptors))
	for _, route := range r.descriptors {
		descriptors = append(descriptors, route)
	}
	return descriptors
}

func (r *Router) setupRoute(route RouteDescriptor, closeSign <-chan struct{}) error {
	var lim iRateLimiter
	var traf iTrafficShapeAlgorithm

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

	pipeline := newRequestPipeline(lim, traf)
	r.setupPath(route.Path, pipeline)
	return nil
}

func (r *RouterBuilder) LoadFromJson(jsonData []byte) error {

	var descriptors []RouteDescriptor

	if err := json.Unmarshal(jsonData, &descriptors); err != nil {
		return fmt.Errorf("falha ao ler JSON: %w", err)
	}

	for _, routeDesc := range descriptors {
		  r.SetRoute(routeDesc);
	}

	return nil
}

func (r *RouterBuilder) LoadFromYaml(yamlData []byte) error {

	var descriptors []RouteDescriptor

	if err := yaml.Unmarshal(yamlData, &descriptors); err != nil {
		return fmt.Errorf("falha ao ler YAML: %w", err)
	}

	for _, routeDesc := range descriptors {
		 r.SetRoute(routeDesc);
	}

	return nil
}

func createTrafficShaperFromDescriptor(strategyDescriptor StrategyDescriptor, closeSign <-chan struct{}) (iTrafficShapeAlgorithm, error) {
	switch strategyDescriptor.StrategyName {
	case TrafficStrategyLeakyBucket:
		params, err := getLeakyBucketTrafficShaperParamsFromMap(strategyDescriptor.Params)
		if err != nil {
			return nil, err
		}
		return newLeakyBucketTrafficShaper(params.Capacity, params.DropPerSecond, closeSign), nil
	default:
		return nil, errors.New("unknown traffic shaper strategy")
	}
}

func createRateLimiterFromDescriptor(routeLimiterDescriptor StrategyDescriptor) (iRateLimiter, error) {

	switch routeLimiterDescriptor.StrategyName {
	case LimiterStrategyFixedWindow:
		params, err := GetFixedWindowRateLimiterParamsFromMap(routeLimiterDescriptor.Params)
		if err != nil {
			return nil, err
		}
		return newFixedWindowRateLimiter(params.Capacity, params.ResetInterval), nil
	case LimiterStrategyTokenBucket:
		params, err := getTokenBucketRateLimiterParamsFromMap(routeLimiterDescriptor.Params)
		if err != nil {
			return nil, err
		}
		return newTokenBucketRateLimiter(params.Capacity, params.RefillRate, params.RequestCost), nil
	case LimiterStrategySlidingWindowLog:
		params, err := getSlidingWindowLogLimiterParamsFromMap(routeLimiterDescriptor.Params)
		if err != nil {
			return nil, err
		}
		return newSlidingWindowLogLimiter(params.Capacity, params.WindowSize), nil
	default:
		return nil, errors.New("unknown limiter strategy")
	}

}
