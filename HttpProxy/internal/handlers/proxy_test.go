package handlers

import (
	"Go_Playground/HttpProxy/internal/items"
	"Go_Playground/HttpProxy/internal/loadbalancer"
	"Go_Playground/HttpProxy/internal/model"
	"Go_Playground/HttpProxy/internal/state"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProxyHandler_Success(t *testing.T) {
	// 1. Create a mock backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock backend response"))
	}))
	defer backend.Close()

	// 2. Setup route in state with healthy backend
	state.Initialize()

	backends := []model.Backend{
		{
			URL:    backend.URL,
			Weight: 1,
			Health: true,
		},
	}

	lb := loadbalancer.NewRoundRobinLoadBalancer(backends)
	routePool := &items.RoutePool{
		Backends:     backends,
		LoadBalancer: lb,
	}

	state.Routes["/test"] = routePool

	// 3. Create a fake request to the proxy
	req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
	w := httptest.NewRecorder()

	// 4. Call the handler
	ProxyHandler(w, req)

	// 5. Verify response
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}
}

func TestProxyHandler_NoMatchingPrefix(t *testing.T) {

	state.Initialize()
	req := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status 502 for unknown prefix, got %d", res.StatusCode)
	}
}

func TestProxyHandler_NoHealthyBackend(t *testing.T) {

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock backend response"))
	}))
	defer backend.Close()

	state.Initialize()

	backends := []model.Backend{
		{
			URL:    backend.URL,
			Weight: 1,
			Health: false,
		},
	}

	lb := loadbalancer.NewRoundRobinLoadBalancer(backends)
	routePool := &items.RoutePool{
		Backends:     backends,
		LoadBalancer: lb,
	}

	state.Routes["/test"] = routePool

	req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status 502 for no healthy backend, got %d", res.StatusCode)
	}
}

func TestProxyHandler_BackendTimeout(t *testing.T) {

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer backend.Close()

	state.Initialize()

	backends := []model.Backend{
		{
			URL:    backend.URL,
			Weight: 1,
			Health: true,
		},
	}

	lb := loadbalancer.NewRoundRobinLoadBalancer(backends)
	routePool := &items.RoutePool{
		Backends:     backends,
		LoadBalancer: lb,
	}

	state.Routes["/test"] = routePool

	req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
	w := httptest.NewRecorder()

	ProxyHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("expected status 504 for backend timeout, got %d", res.StatusCode)
	}
}

func TestProxyHandler_RoundRobin(t *testing.T) {

	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	state.Initialize()

	backends := []model.Backend{
		{
			URL:    backend1.URL,
			Weight: 1,
			Health: true,
		},
		{
			URL:    backend2.URL,
			Weight: 1,
			Health: true,
		},
	}

	lb := loadbalancer.NewRoundRobinLoadBalancer(backends)
	routePool := &items.RoutePool{
		Backends:     backends,
		LoadBalancer: lb,
	}

	state.Routes["/test"] = routePool

	attempts := 10
	for i := range attempts {
		req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
		w := httptest.NewRecorder()

		ProxyHandler(w, req)

		res := w.Result()
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", res.StatusCode)
		}

		if i%2 == 0 && string(body) != "backend1" {
			t.Errorf("expected backend1, got %s", string(body))
		}

		if i%2 == 1 && string(body) != "backend2" {
			t.Errorf("expected backend2, got %s", string(body))
		}
	}
}
