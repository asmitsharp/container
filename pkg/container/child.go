package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type ChildProcess struct{}

func NewChildProcess() *ChildProcess {
	return &ChildProcess{}
}

// run executes the child process with proper namespace setup
func (cp *ChildProcess) Run(args []string) error {
	if err := cp.signalParent(); err != nil {
		return fmt.Errorf("failed to signal parent: %w", err)
	}

	if err := cp.setupContainer(); err != nil {
		return fmt.Errorf("failed to setup container: %w", err)
	}

	return cp.execCommand(args)
}

// signalParent signals the parent process that namespace setup is complete
func (cp *ChildProcess) signalParent() error {
	pipe := os.NewFile(3, "pipe")
	if pipe == nil {
		return fmt.Errorf("communication pipe not found")
	}
	defer pipe.Close()

	if _, err := pipe.Write([]byte{1}); err != nil {
		return fmt.Errorf("failed to write to pipe: %w", err)
	}

	return nil
}

// setupContainer performs container-specific setup
func (cp *ChildProcess) setupContainer() error {
	// Generate unique hostname based on timestamp
	hostname := fmt.Sprintf("container-%d", time.Now().UnixNano())

	// Set hostname using hostname command
	cmd := exec.Command("hostname", hostname)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set hostname: %w", err)
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Remount /proc for PID namespace isolation
	cmd = exec.Command("mount", "-t", "proc", "proc", "/proc")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount /proc: %w", err)
	}

	os.Setenv("PS1", "container# ")
	os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

	return nil
}

func (cp *ChildProcess) execCommand(args []string) error {
	binary, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("executable not found: %s", args[0])
	}

	return syscall.Exec(binary, args, os.Environ())
}
