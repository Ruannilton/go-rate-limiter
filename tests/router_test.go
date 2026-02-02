package tests

import (
	"testing"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func TestRouter_Priorities(t *testing.T) {
	builder  := internal.NewRouterBuilder()
	router := builder.GetRouter()
	closeChan := make(chan struct{})
	defer close(closeChan)

	// Mock descriptors
	staticDesc := internal.RouteDescriptor{
		Path: "/api/v1/users",
		LimiterDescriptor: &internal.StrategyDescriptor{
			StrategyName: internal.LimiterStrategyFixedWindow,
			Params: map[string]any{"capacity": 10, "reset_interval": 60.0},
		},
	}
	varDesc := internal.RouteDescriptor{
		Path: "/api/v1/:id",
		LimiterDescriptor: &internal.StrategyDescriptor{
			StrategyName: internal.LimiterStrategyFixedWindow,
			Params: map[string]any{"capacity": 20, "reset_interval": 60.0},
		},
	}
	wildcardDesc := internal.RouteDescriptor{
		Path: "/api/v1/*",
		LimiterDescriptor: &internal.StrategyDescriptor{
			StrategyName: internal.LimiterStrategyFixedWindow,
			Params: map[string]any{"capacity": 30, "reset_interval": 60.0},
		},
	}

	builder.AddRoute(staticDesc, closeChan)
	builder.AddRoute(varDesc, closeChan)
	builder.AddRoute(wildcardDesc, closeChan)

	// Test Static Match
	_, found := router.EvalRoute("/api/v1/users")
	if !found {
		t.Fatal("Expected to find route /api/v1/users")
	}

	// Test Var Match
	_, found = router.EvalRoute("/api/v1/123")
	if !found {
		t.Fatal("Expected to find route /api/v1/123")
	}
	
	// Test Wildcard Match
	_, found = router.EvalRoute("/api/v1/single-segment")
	if !found {
		t.Fatal("Expected to find route /api/v1/*")
	}

	// Test No Match
	_, found = router.EvalRoute("/api/v2/users")
	if found {
		t.Fatal("Expected NOT to find route /api/v2/users")
	}
}

func TestRouter_ConflictResolution(t *testing.T) {
	builder  := internal.NewRouterBuilder()
	router := internal.NewRouter()
	closeChan := make(chan struct{})
	defer close(closeChan)

	// Add /a (capacity 1)
	builder.AddRoute(internal.RouteDescriptor{
		Path: "/a",
		LimiterDescriptor: &internal.StrategyDescriptor{
			StrategyName: internal.LimiterStrategyFixedWindow,
			Params: map[string]any{"capacity": 1, "reset_interval": 10.0},
		},
	}, closeChan)

	// Add /:b (capacity 100)
	builder.AddRoute(internal.RouteDescriptor{
		Path: "/:b",
		LimiterDescriptor: &internal.StrategyDescriptor{
			StrategyName: internal.LimiterStrategyFixedWindow,
			Params: map[string]any{"capacity": 100, "reset_interval": 10.0},
		},
	}, closeChan)

	// Request /a
	// Should hit Static (/a) -> Capacity 1.
	p, found := router.EvalRoute("/a")
	if !found {
		t.Fatal("Route not found")
	}
	
	// Check capacity by exhausting it.
	// 1st request ok
	resp, _ := p.HandleRequest()
	<-resp.Allowed()

	// 2nd request blocked (if it was capacity 1)
	// If it matched /:b, capacity would be 100, so it would pass.
	resp, _ = p.HandleRequest()
	if allowed := <-resp.Allowed(); allowed {
		t.Error("Expected blocked request, meaning it matched /a (cap 1) not /:b (cap 100)")
	}
}