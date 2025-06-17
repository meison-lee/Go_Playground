package items

import (
	"Go_Playground/HttpProxy/internal/interfaces"
	"Go_Playground/HttpProxy/internal/model"
	"sync"
)

type RoutePool struct {
	Mutex        sync.RWMutex
	Backends     []model.Backend
	LoadBalancer interfaces.LoadBalancer
}

func NewRoutePool(backends []model.Backend, lb interfaces.LoadBalancer) *RoutePool {
	return &RoutePool{
		Backends:     backends,
		LoadBalancer: lb,
	}
}

func (rp *RoutePool) NextBackend() (*model.Backend, error) {
	rp.Mutex.RLock()
	defer rp.Mutex.RUnlock()

	return rp.LoadBalancer.NextBackend(rp.Backends)
}

func (rp *RoutePool) UpdateBackendHealth(index int, health bool) {
	rp.Mutex.Lock()
	defer rp.Mutex.Unlock()

	if index < len(rp.Backends) {
		rp.Backends[index].Health = health
	}
}

func (rp *RoutePool) GetBackends() []model.Backend {
	rp.Mutex.RLock()
	defer rp.Mutex.RUnlock()

	backends := make([]model.Backend, len(rp.Backends))
	copy(backends, rp.Backends)

	return backends
}
