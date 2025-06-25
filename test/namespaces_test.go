//go:build linux

package test

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/ashmitsharp/container/internal/namespaces"
)

// TestNamespaceIsolation tests various namespace isolation features

const rootfsPath = "../busybox-rootfs"

// checkRootfs is a helper to ensure the rootfs for testing exists.
func checkRootfs(t *testing.T) {
	if _, err := os.Stat(rootfsPath); os.IsNotExist(err) {
		t.Fatalf("Test rootfs not found at %s. Please run the setup steps from the documentation.", rootfsPath)
	}
}
func TestNamespaceIsolation(t *testing.T) {
	checkRootfs(t)

	tests := []struct {
		name      string
		flags     namespaces.NamespaceFlags
		validator func(containerPID int) error
	}{
		{
			name:      "UTS namespace isolation",
			flags:     namespaces.NamespaceFlags{UTS: true},
			validator: validateUTSIsolation,
		},
		{
			name:      "PID namespace isolation",
			flags:     namespaces.NamespaceFlags{PID: true},
			validator: validatePIDIsolation,
		},
		{
			name:      "Mount namespace isolation",
			flags:     namespaces.NamespaceFlags{MOUNT: true},
			validator: validateMountIsolation,
		},
		{
			name: "Multiple namespace isolation",
			flags: namespaces.NamespaceFlags{
				UTS:   true,
				PID:   true,
				MOUNT: true,
				IPC:   true,
			},
			validator: validateMultipleNamespaces,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create process executor
			executor := namespaces.NewProcessExecutor(tt.flags)

			// Start container process
			process, err := executor.CreateContainer([]string{"/bin/sleep", "10"}, rootfsPath)
			if err != nil {
				t.Fatalf("Failed to create container: %v", err)
			}

			// Validate isolation
			if err := tt.validator(process.PID); err != nil {
				t.Errorf("Isolation validation failed: %v", err)
			}

			// Clean up
			syscall.Kill(process.PID, syscall.SIGTERM)
		})
	}
}

// validateUTSIsolation checks if UTS namespace isolation works
func validateUTSIsolation(containerPID int) error {
	// Get hostname from container namespace
	cmd := exec.Command("nsenter", "-t", strconv.Itoa(containerPID), "-u", "hostname")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get container hostname: %w", err)
	}

	containerHostname := strings.TrimSpace(string(output))

	// Get host hostname
	hostHostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get host hostname: %w", err)
	}

	// They should be different if UTS namespace is working
	if containerHostname == hostHostname {
		return fmt.Errorf("UTS isolation failed: container hostname (%s) same as host (%s)",
			containerHostname, hostHostname)
	}

	return nil
}

// validatePIDIsolation checks if PID namespace isolation works
func validatePIDIsolation(containerPID int) error {
	// Get process list from container namespace
	cmd := exec.Command("nsenter", "-t", strconv.Itoa(containerPID), "-p", "ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get container process list: %w", err)
	}

	processes := strings.Split(string(output), "\n")

	// Container should have much fewer processes than host
	if len(processes) > 10 {
		return fmt.Errorf("PID isolation questionable: container has %d processes", len(processes))
	}

	return nil
}

// validateMountIsolation checks if mount namespace isolation works
func validateMountIsolation(containerPID int) error {
	// This is a basic check - in a real implementation you would
	// create a mount inside the container and verify it's not visible outside
	cmd := exec.Command("nsenter", "-t", strconv.Itoa(containerPID), "-m", "mount")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get container mounts: %w", err)
	}

	// Container should have its own mount namespace
	if len(string(output)) == 0 {
		return fmt.Errorf("mount namespace appears empty")
	}

	return nil
}

// validateMultipleNamespaces checks if multiple namespaces work together
func validateMultipleNamespaces(containerPID int) error {
	// Run multiple validation checks
	validators := []func(int) error{
		validateUTSIsolation,
		validatePIDIsolation,
		validateMountIsolation,
	}

	for _, validator := range validators {
		if err := validator(containerPID); err != nil {
			return err
		}
	}

	return nil
}

// BenchmarkNamespaceCreation benchmarks namespace creation performance
func BenchmarkNamespaceCreation(b *testing.B) {
	flags := namespaces.DefaultNamespace()
	executor := namespaces.NewProcessExecutor(*flags)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		process, err := executor.CreateContainer([]string{"/bin/true"}, rootfsPath)
		if err != nil {
			b.Fatalf("Failed to create container: %v", err)
		}

		// Wait for process to finish
		syscall.Kill(process.PID, syscall.SIGTERM)
	}
}
