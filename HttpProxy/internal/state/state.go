package state

import (
	"Go_Playground/HttpProxy/internal/types"
	"sync"
	"sync/atomic"
)

// Global state variables
var (
	Routes            map[string]*types.RoutePool
	CountRequest      map[string]int64
	Requests          map[string]types.ProxyRequest
	StatusCodeCount   map[int]int64
	CountRequestMu    sync.Mutex
	RequestsMu        sync.Mutex
	StatusCodeCountMu sync.Mutex
)

var (
	TotalRequests int64
	ErrorCount    int64
)

// Initialize initializes all global state
func Initialize() {
	Routes = make(map[string]*types.RoutePool)
	CountRequest = make(map[string]int64)
	Requests = make(map[string]types.ProxyRequest)
	StatusCodeCount = make(map[int]int64)
}

// Getters and setters for thread-safe access
func AddTotalRequests() int64 {
	return atomic.AddInt64(&TotalRequests, 1)
}

func LoadTotalRequests() int64 {
	return atomic.LoadInt64(&TotalRequests)
}

func AddErrorCount() int64 {
	return atomic.AddInt64(&ErrorCount, 1)
}

func LoadErrorCount() int64 {
	return atomic.LoadInt64(&ErrorCount)
}

func SetRequest(id string, req types.ProxyRequest) {
	RequestsMu.Lock()
	defer RequestsMu.Unlock()
	Requests[id] = req
}

func GetAllRequests() []types.ProxyRequest {
	RequestsMu.Lock()
	defer RequestsMu.Unlock()

	requestsCopy := make([]types.ProxyRequest, 0, len(Requests))
	for _, req := range Requests {
		requestsCopy = append(requestsCopy, req)
	}
	return requestsCopy
}

func IncrementRouteCount(prefix string) {
	CountRequestMu.Lock()
	defer CountRequestMu.Unlock()
	CountRequest[prefix]++
}

func GetRouteCountCopy() map[string]int64 {
	CountRequestMu.Lock()
	defer CountRequestMu.Unlock()

	countRequestCopy := make(map[string]int64, len(CountRequest))
	for k, v := range CountRequest {
		countRequestCopy[k] = v
	}
	return countRequestCopy
}

func IncrementStatusCode(statusCode int) {
	StatusCodeCountMu.Lock()
	defer StatusCodeCountMu.Unlock()
	StatusCodeCount[statusCode]++
}

func GetStatusCodeCountCopy() map[int]int64 {
	StatusCodeCountMu.Lock()
	defer StatusCodeCountMu.Unlock()

	statusCodeCountCopy := make(map[int]int64, len(StatusCodeCount))
	for k, v := range StatusCodeCount {
		statusCodeCountCopy[k] = v
	}
	return statusCodeCountCopy
}
