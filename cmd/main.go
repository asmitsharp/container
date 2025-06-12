package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	// Detect if we're inside the child process
	if len(os.Args) > 1 && os.Args[1] == "child" {
		runContainer()
		return
	}

	// Re-execute self in a new UTS namespace
	cmd := exec.Command("/proc/self/exe", "child")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}

func runContainer() {
	// Set the hostname inside the UTS namespace
	if err := syscall.Sethostname([]byte("mycontainer")); err != nil {
		fmt.Println("Failed to set hostname:", err)
		return
	}

	// Launch shell
	cmd := exec.Command("/bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Shell error:", err)
	}
}
