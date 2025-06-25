//go:build linux

package namespaces

import (
	"fmt"
	"sync"
	"time"
)

type NamespaceManager struct {
	containers map[string]*ContainerProcess
	mutex      sync.RWMutex
}

type ContainerProcess struct {
	PID       int
	Namespace NamespaceFlags
	State     ProcessState
	StartTime time.Time
	ExitCode  int
	Command   []string
}

type ProcessState int

const (
	Created ProcessState = iota
	Running
	Stopped
	Failed
)

func (ps ProcessState) String() string {
	switch ps {
	case Created:
		return "Created"
	case Running:
		return "Running"
	case Stopped:
		return "Stopped"
	case Failed:
		return "Failed"
	default:
		return "Unknown"
	}
}

func NewNameSpaceManager() *NamespaceManager {
	return &NamespaceManager{
		containers: make(map[string]*ContainerProcess),
		mutex:      sync.RWMutex{},
	}
}

func (nm *NamespaceManager) AddContainer(id string, process *ContainerProcess) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.containers[id] = process
}

func (nm *NamespaceManager) GetContainer(id string) (*ContainerProcess, bool) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	container, exists := nm.containers[id]
	return container, exists
}

func (nm *NamespaceManager) UpdateContainerState(id string, state ProcessState, exitCode *int) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	container, exists := nm.containers[id]
	if !exists {
		return fmt.Errorf("container %s not found", id)
	}

	container.State = state
	if exitCode != nil {
		container.ExitCode = *exitCode
	}
	return nil
}

func (nm *NamespaceManager) ListContainer() map[string]*ContainerProcess {
	nm.mutex.RLock()
	defer nm.mutex.Unlock()

	result := make(map[string]*ContainerProcess)
	for id, container := range nm.containers {
		result[id] = container
	}
	return result
}

func (nm *NamespaceManager) RemoveContainer(id string) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	delete(nm.containers, id)
}
