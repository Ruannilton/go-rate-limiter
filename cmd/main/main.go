package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func main() {
	// 1. Create the main router
	router := internal.NewRouter()

	// 2. Setup Fixed Window Algorithm
	fwStorage := internal.NewFixedWindowMemoryStorage()
	fwStorage.SetDefaultValue("/fw", internal.FixedWindowCtrParams{Capacity: 10})
	fwHandler := internal.NewHandler(fwStorage)
	router.AddRoute("/fw", fwHandler)

	// 3. Setup Sliding Window Log Algorithm
	swlStorage := internal.NewSlidingWindowLogMemoryStorage()
	swlStorage.SetDefaultValue("/swl", internal.SlidingWindowLogCtrParams{Capacity: 15, WindowSize: 10 * time.Second})
	swlHandler := internal.NewHandler(swlStorage)
	router.AddRoute("/swl", swlHandler)

	// 4. Setup Leaky Bucket Algorithm
	lbStorage := internal.NewLeakyBucketMemoryStorage()
	lbStorage.SetDefaultValue("/lb", internal.LeakyBucketCtrParams{Capacity: 5, DropPerSecond: 1})
	lbHandler := internal.NewHandler(lbStorage)
	router.AddRoute("/lb", lbHandler)

	// 5. Setup Token Bucket Algorithm
	tbStorage := internal.NewTokenBucketMemoryStorage()
	tbStorage.SetDefaultValue("/tb", internal.TokenBucketCtrParams{Capacity: 20, RefillRate: 2, RequestCost: 1})
	tbHandler := internal.NewHandler(tbStorage)
	router.AddRoute("/tb", tbHandler)

	// 6. Create a simple HTTP server to use the router
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler, found := router.GetRoute(r.URL.Path)
		if !found {
			http.NotFound(w, r)
			return
		}

		// Using a static identifier for demonstration
		identifier := "user-1"
		response, err := handler.Handle(identifier)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if response.IsDelayed() {
			fmt.Fprintf(w, "Request for %s is being processed (delayed).\n", r.URL.Path)
			go func() {
				allowed := response.IsAllowed() // This will block until the channel response
				fmt.Printf("Delayed request for %s (user: %s) was processed. Allowed: %v\n", r.URL.Path, identifier, allowed)
			}()
		} else {
			if response.IsAllowed() {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Request for %s allowed!\n", r.URL.Path)
			} else {
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, "Request for %s denied! Rate limit exceeded.\n", r.URL.Path)
			}
		}
	})

	// 7. Start the server and handle shutdown
	fmt.Println("Server is running on :8080")
	fmt.Println("Try accessing:")
	fmt.Println(" - http://localhost:8080/fw (Fixed Window, 10 requests)")
	fmt.Println(" - http://localhost:8080/swl (Sliding Window, 15 requests per 10s)")
	fmt.Println(" - http://localhost:8080/lb (Leaky Bucket, 1 request per second, capacity 5)")
	fmt.Println(" - http://localhost:8080/tb (Token Bucket, 20 capacity, 2 tokens/sec)")

	server := &http.Server{Addr: ":8080"}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe(): %s\n", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	fmt.Println("\nShutting down server...")
	server.Shutdown(nil)
}