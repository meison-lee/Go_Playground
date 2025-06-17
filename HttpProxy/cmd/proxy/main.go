package main

import (
	"Go_Playground/HttpProxy/internal/config"
	constant "Go_Playground/HttpProxy/internal/constants"
	"Go_Playground/HttpProxy/internal/handlers"
	"Go_Playground/HttpProxy/internal/items"
	"Go_Playground/HttpProxy/internal/loadbalancer"
	"Go_Playground/HttpProxy/internal/model"
	"Go_Playground/HttpProxy/internal/state"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// 1. Load configuration first
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}

	// 2. Initialize global state
	state.Initialize()

	// 3. Populate routes from config
	initializeRoutes()

	// 4. Start health checking for all routes
	startHealthChecking()

	// 5. Setup HTTP server with handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.ProxyHandler)
	mux.HandleFunc("/metrics", handlers.MetricsHandler)
	mux.HandleFunc("/requests", handlers.RequestsHandler)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  constant.ReadTimeout,
		WriteTimeout: constant.WriteTimeout,
		Handler:      mux,
	}

	fmt.Println("Proxy server started at :8080")
	log.Fatal(server.ListenAndServe())
}

func initializeRoutes() {
	for _, route := range config.Config.Routes {
		// Convert config backends to types.Backend
		backends := make([]model.Backend, len(route.Backend))
		for i, backend := range route.Backend {
			backends[i] = model.Backend{
				URL:    backend.URL,
				Weight: backend.Weight,
				Health: true, // Initially assume healthy
			}
		}

		// Create load balancer for this route
		lb := loadbalancer.NewRoundRobinLoadBalancer(backends)

		state.Routes[route.Prefix] = items.NewRoutePool(backends, lb)
	}
}

func startHealthChecking() {
	for _, routePool := range state.Routes {
		go func(routePool *items.RoutePool) {
			for {
				backends := routePool.GetBackends()
				for i, backend := range backends {
					client := http.Client{
						Timeout: 2 * time.Second,
					}
					resp, err := client.Get(backend.URL)
					health := err == nil && resp.StatusCode == 200
					if resp != nil {
						resp.Body.Close()
					}
					routePool.UpdateBackendHealth(i, health)
				}
				time.Sleep(5 * time.Second)
			}
		}(routePool)
	}
}
