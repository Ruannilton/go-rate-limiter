package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func main() {
	// 1. Create the main builder
	builder := internal.NewRouterBuilder()
	closeChannel := make(chan struct{})
	defer func(){
		closeChannel <- struct{}{}
		close(closeChannel)
	}()

	data, err := os.ReadFile("routes.json")
    if err != nil {
        log.Fatal(err)
    }
	err = builder.LoadFromJson(data, closeChannel)
	if err != nil {
		log.Fatal(err)
	}
	
	router := builder.GetRouter()
	
	for i := 0; i < 20; i++ {
		pipeline, ok :=router.EvalRoute("/api/v1/users")
		if !ok {
			fmt.Println("Route not found")
			continue
		}
	
		
		go func(){
				response, err := pipeline.HandleRequest()
				if err != nil {
					fmt.Println("Error handling request:", err)
					return
				}
				allowed := <-response.Allowed()
		
		fmt.Printf("Fixed Window Request %d allowed: %v\n", i+1, allowed)
		}()

	
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	fmt.Println("\nShutting down server...")
	
}