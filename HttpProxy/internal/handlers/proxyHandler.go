package handlers

import (
	"Go_Playground/HttpProxy/internal/constant"
	"Go_Playground/HttpProxy/internal/state"
	"Go_Playground/HttpProxy/internal/types"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	recorder := &StatusRecorder{
		ResponseWriter: w,
		StatusCode:     200,
	}

	state.AddTotalRequests()

	request := types.ProxyRequest{
		Path:      r.URL.Path,
		StartTime: start,
		RequestID: fmt.Sprintf("%d", state.LoadTotalRequests()),
	}

	state.SetRequest(request.RequestID, request)

	for prefix, routePool := range state.Routes {
		if strings.HasPrefix(r.URL.Path, prefix) {
			var selectedBackend types.Backend

			selectedBackend, err := routePool.LoadBalancer.NextBackend()
			if err != nil {
				http.Error(w, "No healthy backend available", http.StatusBadGateway)
				return
			}

			state.IncrementRouteCount(prefix)

			request.Backend = selectedBackend.URL
			target, _ := url.Parse(selectedBackend.URL)
			proxy := httputil.NewSingleHostReverseProxy(target)

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), constant.ProxyTimeout)
			defer cancel()

			// Create a new request with the timeout context
			proxyReq := r.WithContext(ctx)

			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				state.AddErrorCount()
				request.Error = err
				if err == context.DeadlineExceeded {
					request.StatusCode = http.StatusGatewayTimeout
					state.IncrementStatusCode(http.StatusGatewayTimeout)
					http.Error(w, "Proxy Timeout", http.StatusGatewayTimeout)
				} else {
					request.StatusCode = http.StatusBadGateway
					state.IncrementStatusCode(http.StatusBadGateway)
					http.Error(w, "Proxy Error", http.StatusBadGateway)
				}
				state.SetRequest(request.RequestID, request)
				log.Printf("Proxy error: %v", err)
			}

			proxy.ServeHTTP(recorder, proxyReq)
			request.Duration = time.Since(start)
			request.StatusCode = recorder.StatusCode
			state.IncrementStatusCode(recorder.StatusCode)

			state.SetRequest(request.RequestID, request)
			log.Printf("Proxy %s -> %s, cost %v\n", r.URL.Path, selectedBackend.URL, time.Since(start))

			return
		}
	}
}

// func metricsHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/plain")
// 	fmt.Fprintf(w, "proxy_requests_total %d\n", atomic.LoadInt64(&totalRequests))
// 	fmt.Fprintf(w, "proxy_errors_total %d\n", atomic.LoadInt64(&errorCount))

// 	// Copy countRequest data
// 	countRequestMu.Lock()
// 	countRequestCopy := make(map[string]int64, len(countRequest))
// 	maps.Copy(countRequestCopy, countRequest)
// 	countRequestMu.Unlock()

// 	// Copy statusCodeCount data
// 	statusCodeCountMu.Lock()
// 	statusCodeCountCopy := make(map[int]int64, len(statusCodeCount))
// 	for k, v := range statusCodeCount {
// 		statusCodeCountCopy[k] = v
// 	}
// 	statusCodeCountMu.Unlock()

// 	// Process the copied data without holding locks
// 	for prefix, count := range countRequestCopy {
// 		fmt.Fprintf(w, "proxy_requests_%s %d\n", prefix[1:len(prefix)-1], count)
// 	}
// 	for statusCode, count := range statusCodeCountCopy {
// 		fmt.Fprintf(w, "proxy_status_code_total{code=\"%d\"} %d\n", statusCode, count)
// 	}
// 	for prefix, routePool := range routes {
// 		routePool.mutex.Lock()
// 		fmt.Fprintf(w, "\nPrefix:%s\n", prefix)
// 		for _, backend := range routePool.Backends {
// 			fmt.Fprintf(w, "backend_health{prefix=\"%s\",url=\"%s\"} %t\n", prefix, backend.URL, backend.Health)
// 		}
// 		routePool.mutex.Unlock()
// 	}
// }

// func requestsHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	output := RequestsReponse{
// 		TotalRequests: atomic.LoadInt64(&totalRequests),
// 		TotalErrors:   atomic.LoadInt64(&errorCount),
// 	}

// 	// Copy requests data
// 	requestsMu.Lock()
// 	requestsCopy := make([]ProxyRequest, 0, len(requests))
// 	for _, req := range requests {
// 		requestsCopy = append(requestsCopy, req)
// 	}
// 	requestsMu.Unlock()

// 	// Process the copied data without holding the lock
// 	output.Requests = requestsCopy
// 	json.NewEncoder(w).Encode(output)
// }
