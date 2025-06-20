package runtime

import (
	"crypto/rand"
	"fmt"
	"os"
	"syscall"

	"github.com/ashmitsharp/container/internal/namespaces"
	"github.com/ashmitsharp/container/pkg/container"
)

// manages container life-cycle
type RuntimeManager struct {
	namespaceManager *namespaces.NamespaceManager
	containers       map[string]*RunningContainer
}

type RunningContainer struct {
	ID      string
	Config  *container.ContainerConfig
	Process *namespaces.ContainerProcess
	Cmd     *os.Process
}

func NewRunTimeManager() *RuntimeManager {
	return &RuntimeManager{
		namespaceManager: namespaces.NewNameSpaceManager(),
		containers:       make(map[string]*RunningContainer),
	}
}

func (rm *RuntimeManager) CreateContainer(config *container.ContainerConfig) (string, error) {
	containerID := rm.generateContainerID()

	executor := namespaces.NewProcessExecutor(config.Namespaces)

	process, err := executor.CreateContainer(config.Command, config.Rootfs)
	if err != nil {
		return "", fmt.Errorf("failed to create container process: %w", err)
	}

	runningContainer := &RunningContainer{
		ID:      containerID,
		Config:  config,
		Process: process,
		Cmd:     &os.Process{Pid: process.PID},
	}

	rm.containers[containerID] = runningContainer
	rm.namespaceManager.AddContainer(containerID, process)

	return containerID, nil
}

func (rm *RuntimeManager) StartContainer(containerID string) error {
	container, exists := rm.containers[containerID]
	if !exists {
		return fmt.Errorf("container %s not found", containerID)
	}

	fmt.Printf("Container %s started with PID %d\n", containerID, container.Process.PID)
	return nil
}

func (rm *RuntimeManager) WaitContainer(containerID string) error {
	container, exists := rm.containers[containerID]
	if !exists {
		return fmt.Errorf("container %s not found", containerID)
	}

	processState, err := container.Cmd.Wait()
	if err != nil {
		return fmt.Errorf("error waiting for container: %w", err)
	}

	exitCode := 0
	if processState != nil {
		if exitStatus, ok := processState.Sys().(syscall.WaitStatus); ok {
			exitCode = exitStatus.ExitStatus()
		}
	}

	var finalState namespaces.ProcessState
	if exitCode == 0 {
		finalState = namespaces.Stopped
	} else {
		finalState = namespaces.Failed
	}

	rm.namespaceManager.UpdateContainerState(containerID, finalState, &exitCode)

	fmt.Printf("Container %s finished with exit code %d\n", containerID, exitCode)
	return nil
}

func (rm *RuntimeManager) generateContainerID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("container-%x", bytes)
}
