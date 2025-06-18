package main

import (
	"fmt"
	"os"

	"github.com/ashmitsharp/container/pkg/container"
	"github.com/ashmitsharp/container/pkg/runtime"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <command> [args...]\n", os.Args[0])
		fmt.Println("Commands:")
		fmt.Println("  run <image> [command]  - Run a container")
		fmt.Println("  child [command]        - Internal child process (don't call directly)")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: run <image> [command]")
			os.Exit(1)
		}
		handleRun(os.Args[2:])
	case "child":
		if len(os.Args) < 3 {
			fmt.Println("No command specified for child")
			os.Exit(1)
		}
		handleChild(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func handleRun(args []string) {
	config := &container.ContainerConfig{
		Image:   args[0],
		Command: []string{"/bin/sh"},
	}

	if len(args) > 1 {
		config.Command = args[1:]
	}

	runtime := runtime.NewRunTimeManager()

	containerID, err := runtime.CreateContainer(config)
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting container %s...\n", containerID)

	if err := runtime.StartContainer(containerID); err != nil {
		fmt.Printf("Failed to start container: %v\n", err)
		os.Exit(1)
	}

	// Wait for container to finish
	if err := runtime.WaitContainer(containerID); err != nil {
		fmt.Printf("Container error: %v\n", err)
		os.Exit(1)
	}
}

func handleChild(args []string) {
	child := container.NewChildProcess()
	if err := child.Run(args); err != nil {
		fmt.Printf("Child process failed: %v\n", err)
		os.Exit(1)
	}
}
