package rate_limiter

import (
	"testing"
)

func TestRouter_Priorities(t *testing.T) {
	builder := NewRouterBuilder()
	router := builder.GetRouter()
	closeChan := make(chan struct{})
	defer close(closeChan)

	// Mock descriptors
	staticDesc := RouteDescriptor{
		Path: "/api/v1/users",
		LimiterDescriptor: &StrategyDescriptor{
			StrategyName: LimiterStrategyFixedWindow,
			Params:       map[string]any{"capacity": 10, "reset_interval": 60.0},
		},
	}
	varDesc := RouteDescriptor{
		Path: "/api/v1/:id",
		LimiterDescriptor: &StrategyDescriptor{
			StrategyName: LimiterStrategyFixedWindow,
			Params:       map[string]any{"capacity": 20, "reset_interval": 60.0},
		},
	}
	wildcardDesc := RouteDescriptor{
		Path: "/api/v1/*",
		LimiterDescriptor: &StrategyDescriptor{
			StrategyName: LimiterStrategyFixedWindow,
			Params:       map[string]any{"capacity": 30, "reset_interval": 60.0},
		},
	}

	builder.AddRoute(staticDesc, closeChan)
	builder.AddRoute(varDesc, closeChan)
	builder.AddRoute(wildcardDesc, closeChan)

	// Test Static Match
	_, found := router.evalRoute("/api/v1/users")
	if !found {
		t.Fatal("Expected to find route /api/v1/users")
	}

	// Test Var Match
	_, found = router.evalRoute("/api/v1/123")
	if !found {
		t.Fatal("Expected to find route /api/v1/123")
	}

	// Test Wildcard Match
	_, found = router.evalRoute("/api/v1/single-segment")
	if !found {
		t.Fatal("Expected to find route /api/v1/*")
	}

	// Test No Match
	_, found = router.evalRoute("/api/v2/users")
	if found {
		t.Fatal("Expected NOT to find route /api/v2/users")
	}
}

func TestRouter_ConflictResolution(t *testing.T) {
	builder := NewRouterBuilder()
	router := NewRouter()
	closeChan := make(chan struct{})
	defer close(closeChan)

	// Add /a (capacity 1)
	builder.AddRoute(RouteDescriptor{
		Path: "/a",
		LimiterDescriptor: &StrategyDescriptor{
			StrategyName: LimiterStrategyFixedWindow,
			Params:       map[string]any{"capacity": 1, "reset_interval": 10.0},
		},
	}, closeChan)

	// Add /:b (capacity 100)
	builder.AddRoute(RouteDescriptor{
		Path: "/:b",
		LimiterDescriptor: &StrategyDescriptor{
			StrategyName: LimiterStrategyFixedWindow,
			Params:       map[string]any{"capacity": 100, "reset_interval": 10.0},
		},
	}, closeChan)

	// Request /a
	// Should hit Static (/a) -> Capacity 1.
	p, found := router.evalRoute("/a")
	if !found {
		t.Fatal("Route not found")
	}

	// Check capacity by exhausting it.
	// 1st request ok
	resp := p.handleRequest()
	<-resp.Allowed()

	// 2nd request blocked (if it was capacity 1)
	// If it matched /:b, capacity would be 100, so it would pass.
	resp = p.handleRequest()
	if allowed := <-resp.Allowed(); allowed {
		t.Error("Expected blocked request, meaning it matched /a (cap 1) not /:b (cap 100)")
	}
}
