package rate_limiter

import (
	"testing"
)

func TestRouter_Priorities(t *testing.T) {
	closeChan := make(chan struct{})
	defer close(closeChan)
	
	builder := NewRouterBuilder(closeChan)

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

	builder.SetRoute(staticDesc)
	builder.SetRoute(varDesc)
	builder.SetRoute(wildcardDesc)

	router := builder.Build()

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
	closeChan := make(chan struct{})
	defer close(closeChan)

	builder := NewRouterBuilder(closeChan)

	// Add /a (capacity 1)
	builder.SetRoute(RouteDescriptor{
		Path: "/a",
		LimiterDescriptor: &StrategyDescriptor{
			StrategyName: LimiterStrategyFixedWindow,
			Params:       map[string]any{"capacity": 1, "reset_interval": 10.0},
		},
	})

	// Add /:b (capacity 100)
	builder.SetRoute(RouteDescriptor{
		Path: "/:b",
		LimiterDescriptor: &StrategyDescriptor{
			StrategyName: LimiterStrategyFixedWindow,
			Params:       map[string]any{"capacity": 100, "reset_interval": 10.0},
		},
	})

	router := builder.Build()

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

	

	func TestRouterBuilder_Export(t *testing.T) {

		closeChan := make(chan struct{})

		defer close(closeChan)

		builder := NewRouterBuilder(closeChan)

	

		desc := RouteDescriptor{

			Path: "/api/test",

			LimiterDescriptor: &StrategyDescriptor{

				StrategyName: LimiterStrategyFixedWindow,

				Params:       map[string]any{"capacity": 10.0, "reset_interval": 60.0},

			},

		}

		builder.SetRoute(desc)

	

		// Test JSON Export

		jsonData, err := builder.ExportToJson()

		if err != nil {

			t.Fatalf("Failed to export to JSON: %v", err)

		}

	

		// Try to reload to verify validity

		builder2 := NewRouterBuilder(closeChan)

		if err := builder2.LoadFromJson(jsonData); err != nil {

			t.Fatalf("Failed to reload exported JSON: %v", err)

		}

		if len(builder2.GetRouteDescriptors()) != 1 {

			t.Errorf("Expected 1 descriptor after reload, got %d", len(builder2.GetRouteDescriptors()))

		}

	

		// Test YAML Export

		yamlData, err := builder.ExportToYaml()

		if err != nil {

			t.Fatalf("Failed to export to YAML: %v", err)

		}

	

		// Try to reload to verify validity

		builder3 := NewRouterBuilder(closeChan)

		if err := builder3.LoadFromYaml(yamlData); err != nil {

			t.Fatalf("Failed to reload exported YAML: %v", err)

		}

		if len(builder3.GetRouteDescriptors()) != 1 {

			t.Errorf("Expected 1 descriptor after reload, got %d", len(builder3.GetRouteDescriptors()))

		}

	}

	