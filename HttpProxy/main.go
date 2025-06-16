package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/exp/maps"
)

type ProxyConfig struct {
	Routes []RouteConfig `json:"routes"`
}

type RouteConfig struct {
	Prefix  string    `json:"prefix"`
	Backend []Backend `json:"backends"`
}

type Backend struct {
	URL    string `json:"url"`
	Weight int64  `json:"weight"`
	Health bool   `json:"health"`
}
type RequestsReponse struct {
	TotalRequests int64          `json:"total_requests"`
	TotalErrors   int64          `json:"total_errors"`
	Requests      []ProxyRequest `json:"requests"`
}

type RoutePool struct {
	Index    int64
	mutex    sync.Mutex
	Backends []Backend
}

type ProxyRequest struct {
	Path       string
	Backend    string
	StartTime  time.Time
	Duration   time.Duration
	Error      error
	StatusCode int
	RequestID  string
}

var (
	routes            map[string]*RoutePool
	countRequest      map[string]int64
	requests          map[string]ProxyRequest
	statusCodeCount   map[int]int64
	countRequestMu    sync.Mutex
	requestsMu        sync.Mutex
	statusCodeCountMu sync.Mutex
)

var totalRequests int64
var errorCount int64

// Timeout configuration
const (
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
	proxyTimeout = 30 * time.Second
)

func LoadConfig(path string) (*ProxyConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ProxyConfig

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func main() {
	configPtr, err := LoadConfig("./config.json")
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}

	routes = make(map[string]*RoutePool)
	for _, route := range configPtr.Routes {
		routes[route.Prefix] = &RoutePool{
			Backends: route.Backend,
		}
	}
	countRequest = make(map[string]int64)
	requests = make(map[string]ProxyRequest)
	statusCodeCount = make(map[int]int64)

	// Create a new ServeMux for routing
	mux := http.NewServeMux()
	mux.HandleFunc("/", proxyHandler)
	mux.HandleFunc("/metrics", metricsHandler)
	mux.HandleFunc("/requests", requestsHandler)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Handler:      mux,
	}

	for _, routePool := range routes {
		go func(routePool *RoutePool) {
			for {
				routePool.mutex.Lock()
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
				routePool.mutex.Unlock()
				time.Sleep(5 * time.Second)
			}
		}(routePool)
	}

	fmt.Println("Proxy server started at :8080")
	log.Fatal(server.ListenAndServe())
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	recorder := &StatusRecorder{
		ResponseWriter: w,
		StatusCode:     200,
	}

	start := time.Now()
	atomic.AddInt64(&totalRequests, 1)

	request := ProxyRequest{
		Path:      r.URL.Path,
		StartTime: start,
		RequestID: fmt.Sprintf("%d", atomic.LoadInt64(&totalRequests)),
	}

	requestsMu.Lock()
	requests[request.RequestID] = request
	requestsMu.Unlock()

	for prefix, routePool := range routes {
		if strings.HasPrefix(r.URL.Path, prefix) {
			var selectedBackend Backend

			healthBackendFound := false

			routePool.mutex.Lock()
			maxAttempts := len(routePool.Backends)
			for i := 0; i < maxAttempts; i++ {
				selectedBackend = routePool.Backends[routePool.Index]
				routePool.Index = (routePool.Index + 1) % int64(len(routePool.Backends))

				if selectedBackend.Health {
					healthBackendFound = true
					break
				}
			}
			routePool.mutex.Unlock()

			if !healthBackendFound {
				http.Error(w, "No healthy backend available", http.StatusBadGateway)
				return
			}

			countRequestMu.Lock()
			countRequest[prefix]++
			countRequestMu.Unlock()

			request.Backend = selectedBackend.URL
			target, _ := url.Parse(selectedBackend.URL)
			proxy := httputil.NewSingleHostReverseProxy(target)

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), proxyTimeout)
			defer cancel()

			// Create a new request with the timeout context
			proxyReq := r.WithContext(ctx)

			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				atomic.AddInt64(&errorCount, 1)
				request.Error = err
				if err == context.DeadlineExceeded {
					request.StatusCode = http.StatusGatewayTimeout
					statusCodeCountMu.Lock()
					statusCodeCount[http.StatusGatewayTimeout]++
					statusCodeCountMu.Unlock()
					http.Error(w, "Proxy Timeout", http.StatusGatewayTimeout)
				} else {
					request.StatusCode = http.StatusBadGateway
					statusCodeCountMu.Lock()
					statusCodeCount[http.StatusBadGateway]++
					statusCodeCountMu.Unlock()
					http.Error(w, "Proxy Error", http.StatusBadGateway)
				}
				requestsMu.Lock()
				requests[request.RequestID] = request
				requestsMu.Unlock()
				log.Printf("Proxy error: %v", err)
			}

			proxy.ServeHTTP(recorder, proxyReq)
			request.Duration = time.Since(start)
			request.StatusCode = recorder.StatusCode
			statusCodeCountMu.Lock()
			statusCodeCount[recorder.StatusCode]++
			statusCodeCountMu.Unlock()

			requestsMu.Lock()
			requests[request.RequestID] = request
			requestsMu.Unlock()
			log.Printf("Proxy %s -> %s, cost %v\n", r.URL.Path, selectedBackend.URL, time.Since(start))

			return
		}
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "proxy_requests_total %d\n", atomic.LoadInt64(&totalRequests))
	fmt.Fprintf(w, "proxy_errors_total %d\n", atomic.LoadInt64(&errorCount))

	// Copy countRequest data
	countRequestMu.Lock()
	countRequestCopy := make(map[string]int64, len(countRequest))
	maps.Copy(countRequestCopy, countRequest)
	countRequestMu.Unlock()

	// Copy statusCodeCount data
	statusCodeCountMu.Lock()
	statusCodeCountCopy := make(map[int]int64, len(statusCodeCount))
	for k, v := range statusCodeCount {
		statusCodeCountCopy[k] = v
	}
	statusCodeCountMu.Unlock()

	// Process the copied data without holding locks
	for prefix, count := range countRequestCopy {
		fmt.Fprintf(w, "proxy_requests_%s %d\n", prefix[1:len(prefix)-1], count)
	}
	for statusCode, count := range statusCodeCountCopy {
		fmt.Fprintf(w, "proxy_status_code_total{code=\"%d\"} %d\n", statusCode, count)
	}
	for prefix, routePool := range routes {
		routePool.mutex.Lock()
		fmt.Fprintf(w, "\nPrefix:%s\n", prefix)
		for _, backend := range routePool.Backends {
			fmt.Fprintf(w, "backend_health{prefix=\"%s\",url=\"%s\"} %t\n", prefix, backend.URL, backend.Health)
		}
		routePool.mutex.Unlock()
	}
}

func requestsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	output := RequestsReponse{
		TotalRequests: atomic.LoadInt64(&totalRequests),
		TotalErrors:   atomic.LoadInt64(&errorCount),
	}

	// Copy requests data
	requestsMu.Lock()
	requestsCopy := make([]ProxyRequest, 0, len(requests))
	for _, req := range requests {
		requestsCopy = append(requestsCopy, req)
	}
	requestsMu.Unlock()

	// Process the copied data without holding the lock
	output.Requests = requestsCopy
	json.NewEncoder(w).Encode(output)
}
