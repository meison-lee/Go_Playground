package interfaces

import (
	"Go_Playground/HttpProxy/internal/model"
)

type LoadBalancer interface {
	NextBackend([]model.Backend) (*model.Backend, error)
}
