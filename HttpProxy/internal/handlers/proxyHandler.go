package handlers

import (
	constant "Go_Playground/HttpProxy/internal/constants"
	"Go_Playground/HttpProxy/internal/model"
	"Go_Playground/HttpProxy/internal/state"
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

	request := model.ProxyRequest{
		Path:      r.URL.Path,
		StartTime: start,
		RequestID: fmt.Sprintf("%d", state.LoadTotalRequests()),
	}

	state.SetRequest(request.RequestID, request)

	for prefix, routePool := range state.Routes {
		if strings.HasPrefix(r.URL.Path, prefix) {

			selectedBackend, err := routePool.NextBackend()
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
	request.StatusCode = http.StatusBadGateway
	state.IncrementStatusCode(http.StatusBadGateway)
	http.Error(w, "No matching prefix", http.StatusBadGateway)
	state.SetRequest(request.RequestID, request)
}
