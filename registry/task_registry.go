package registry

import (
	"sync"

	"github.com/shashank-mugiwara/joyboy/task"
)

type ServiceRegistry struct {
	Registry map[string]interface{}
	Mu       sync.RWMutex
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		Registry: make(map[string]interface{}),
	}
}

func (sr *ServiceRegistry) UpdateRegistry(tasks []task.Task) {
}
