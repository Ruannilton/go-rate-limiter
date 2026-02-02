# Go Rate Limiter

A flexible and extensible rate limiting and traffic shaping library for Go, supporting multiple strategies and path-based configuration (including wildcards and variables).

## Features

- **Multiple Rate Limiting Strategies:**
  - **Fixed Window:** Simple counting per time interval.
  - **Token Bucket:** Allows for bursts of traffic with a steady refill rate.
  - **Sliding Window Log:** Precise limiting based on a moving time window.
- **Traffic Shaping:**
  - **Leaky Bucket:** Smooths out traffic spikes by processing requests at a constant rate.
- **Dynamic Routing:**
  - Support for static paths (`/api/v1/users`).
  - Support for URL variables (`/api/:id`).
  - Support for wildcards (`/api/*`).
- **Flexible Configuration:** Load routes and limits from JSON, YAML, or directly via code.

## Installation

```bash
go get github.com/Ruannilton/go-rate-limiter
```

## Quick Start

### 1. Basic Setup via Code

```go
package main

import (
	"fmt"
	"github.com/Ruannilton/go-rate-limiter"
)

func main() {
	closeChan := make(chan struct{})
	defer close(closeChan)

	// Create builder with the close channel
	builder := rate_limiter.NewRouterBuilder(closeChan)

	// Configure a route with a Fixed Window limiter
	builder.SetRoute(rate_limiter.RouteDescriptor{
		Path: "/api/v1/users",
		LimiterDescriptor: &rate_limiter.StrategyDescriptor{
			StrategyName: rate_limiter.LimiterStrategyFixedWindow,
			Params: map[string]any{
				"capacity":       10,    // 10 requests
				"reset_interval": 60.0,  // per 60 seconds
			},
		},
	})

	// Build the router
	router := builder.Build()

	// Evaluate a request
	resp, found := router.HandleRequest("/api/v1/users")
	if found {
		if <-resp.Allowed() {
			fmt.Println("Request allowed!")
		} else {
			fmt.Println("Request throttled.")
		}
	}
}
```

### 2. Loading from JSON/YAML

```go
jsonData := []byte(`[
    {
        "path": "/api/:id",
        "limiter": {
            "type": "token_bucket",
            "params": {
                "capacity": 20,
                "refill_rate": 2,
                "request_cost": 1
            }
        }
    }
]`)

closeChan := make(chan struct{})
defer close(closeChan)

builder := rate_limiter.NewRouterBuilder(closeChan)
builder.LoadFromJson(jsonData)
router := builder.Build()
```

## Core Components

### RouterBuilder
The primary way to configure the library.
- `NewRouterBuilder(<-chan struct{})`: Creates a new builder instance.
- `SetRoute(RouteDescriptor)`: Adds or updates a single route configuration.
- `LoadFromJson([]byte)`: Batches routes from JSON.
- `LoadFromYaml([]byte)`: Batches routes from YAML.
- `Build()`: Finalizes configuration and returns the `Router`.

### Router
Used at runtime to match paths and evaluate limits.
- `HandleRequest(path string) (RequestPipelineResponse, bool)`: Returns the evaluation result and whether the path matched a configured route.

### RequestPipelineResponse
Handles the result of an evaluation, abstracting the difference between an immediate block/allow and a queued request (traffic shaping).
- `Allowed() <-chan bool`: Returns a channel that yields `true` when the request can proceed or `false` if rejected.
- `IsAsync() bool`: Returns `true` if the request was handled by a traffic shaper (e.g., Leaky Bucket) and might have been delayed.

## Strategy Parameters

### `fixed_window`
- `capacity` (float64): Max requests per window.
- `reset_interval` (float64): Window size in seconds.

### `token_bucket`
- `capacity` (float64): Max tokens in bucket.
- `refill_rate` (float64): Tokens added per second.
- `request_cost` (float64): Tokens consumed per request.

### `sliding_window_log`
- `capacity` (int): Max requests in the window.
- `window_size` (float64): Window size in seconds.

### `leaky_bucket` (Traffic Shaper)
- `capacity` (int): Queue size.
- `drop_per_second` (int): How many requests are processed per second.

## License

[MIT](LICENSE)