package types

import (
	"sync"
	"time"
)

type Backend struct {
	URL    string `json:"url"`
	Weight int64  `json:"weight"`
	Health bool   `json:"health"`
}

// LoadBalancer interface to make it testable and extensible
type LoadBalancer interface {
	NextBackend() (Backend, error)
	GetBackends() []Backend
}

type RoutePool struct {
	Index        int64
	Mutex        sync.Mutex
	Backends     []Backend
	LoadBalancer LoadBalancer
}

// Request and response types
type ProxyRequest struct {
	Path       string        `json:"path"`
	Backend    string        `json:"backend"`
	StartTime  time.Time     `json:"start_time"`
	Duration   time.Duration `json:"duration"`
	Error      error         `json:"error,omitempty"`
	StatusCode int           `json:"status_code"`
	RequestID  string        `json:"request_id"`
}

type RequestsResponse struct {
	TotalRequests int64          `json:"total_requests"`
	TotalErrors   int64          `json:"total_errors"`
	Requests      []ProxyRequest `json:"requests"`
}
