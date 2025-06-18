package namespaces

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type ProcessExecutor struct {
	config NamespaceFlags
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func NewProcessExecutor(config NamespaceFlags) *ProcessExecutor {
	return &ProcessExecutor{
		config: config,
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (pe *ProcessExecutor) SetIO(stdin io.Reader, stdout, stderr io.Writer) {
	pe.stdin = stdin
	pe.stdout = stdout
	pe.stderr = stderr
}

func (pe *ProcessExecutor) CreateContainer(command []string) (*ContainerProcess, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	parentRead, childWrite, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create communication pipe: %w", err)
	}
	defer parentRead.Close()

	//command to execute the child process
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, command...)...)

	// Set namespace flags for clone
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: pe.config.ToCloneFlags(),
	}

	// Set up process I/O streams
	cmd.Stdin = pe.stdin
	cmd.Stdout = pe.stdout
	cmd.Stderr = pe.stderr

	// passing communication pipe to child process
	cmd.ExtraFiles = []*os.File{childWrite}

	if err := cmd.Start(); err != nil {
		childWrite.Close()
		return nil, fmt.Errorf("failed to start container process: %w", err)
	}

	childWrite.Close()

	setupResult := make([]byte, 1)
	if _, err := parentRead.Read(setupResult); err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return nil, fmt.Errorf("child process setup failed: %w", err)
	}

	container := &ContainerProcess{
		PID:       cmd.Process.Pid,
		Namespace: pe.config,
		State:     Running,
		StartTime: time.Now(),
		Command:   command,
	}

	return container, nil
}
