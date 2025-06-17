package main

import (
	"Go_Playground/HttpProxy/internal/config"
	"Go_Playground/HttpProxy/internal/constant"
	"Go_Playground/HttpProxy/internal/handlers"
	"Go_Playground/HttpProxy/internal/loadbalancer"
	"Go_Playground/HttpProxy/internal/state"
	"Go_Playground/HttpProxy/internal/types"
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
		backends := make([]types.Backend, len(route.Backend))
		for i, backend := range route.Backend {
			backends[i] = types.Backend{
				URL:    backend.URL,
				Weight: backend.Weight,
				Health: true, // Initially assume healthy
			}
		}

		// Create load balancer for this route
		lb := loadbalancer.NewRoundRobinLoadBalancer(backends)

		state.Routes[route.Prefix] = &types.RoutePool{
			Backends:     backends,
			LoadBalancer: lb,
		}
	}
}

func startHealthChecking() {
	for _, routePool := range state.Routes {
		go func(routePool *types.RoutePool) {
			for {
				routePool.Mutex.Lock()
				for i := range routePool.Backends {
					backend := &routePool.Backends[i]
					client := http.Client{
						Timeout: 2 * time.Second,
					}
					resp, err := client.Get(backend.URL)
					if err == nil && resp.StatusCode == 200 {
						backend.Health = true
					} else {
						backend.Health = false
					}
					if resp != nil {
						resp.Body.Close()
					}
				}
				routePool.Mutex.Unlock()
				time.Sleep(5 * time.Second)
			}
		}(routePool)
	}
}
