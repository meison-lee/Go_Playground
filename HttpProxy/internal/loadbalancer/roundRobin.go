package loadbalancer

import (
	"errors"

	"Go_Playground/HttpProxy/internal/interfaces"
	"Go_Playground/HttpProxy/internal/model"
)

type RoundRobinLoadBalancer struct {
	currentIndex int
}

func NewRoundRobinLoadBalancer(backends []model.Backend) interfaces.LoadBalancer {
	return &RoundRobinLoadBalancer{currentIndex: 0}
}

func (lb *RoundRobinLoadBalancer) NextBackend(backends []model.Backend) (*model.Backend, error) {

	healthBackendFound := false
	var selectedBackend *model.Backend

	maxAttempts := len(backends)
	for i := 0; i < maxAttempts; i++ {
		selectedBackend = &backends[lb.currentIndex]
		lb.currentIndex = (lb.currentIndex + 1) % len(backends)

		if selectedBackend.Health {
			healthBackendFound = true
			break
		}
	}

	if !healthBackendFound {
		return &model.Backend{}, errors.New("no healthy backend available")
	}

	return selectedBackend, nil
}
