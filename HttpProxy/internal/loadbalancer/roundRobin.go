package loadbalancer

import (
	"errors"
	"sync"

	"Go_Playground/HttpProxy/internal/types"
)

type RoundRobinLoadBalancer struct {
	backends     []types.Backend
	currentIndex int
	mutex        sync.Mutex
}

func NewRoundRobinLoadBalancer(backends []types.Backend) types.LoadBalancer {
	return &RoundRobinLoadBalancer{backends: backends, currentIndex: 0}
}

func (lb *RoundRobinLoadBalancer) NextBackend() (types.Backend, error) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	healthBackendFound := false
	var selectedBackend types.Backend

	maxAttempts := len(lb.backends)
	for i := 0; i < maxAttempts; i++ {
		selectedBackend = lb.backends[lb.currentIndex]
		lb.currentIndex = (lb.currentIndex + 1) % len(lb.backends)

		if selectedBackend.Health {
			healthBackendFound = true
			break
		}
	}

	if !healthBackendFound {
		return types.Backend{}, errors.New("no healthy backend available")
	}

	return selectedBackend, nil
}

func (lb *RoundRobinLoadBalancer) GetBackends() []types.Backend {
	return lb.backends
}
