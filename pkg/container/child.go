// package container

// import (
// 	"fmt"
// 	"os"
// 	"os/exec"
// 	"syscall"
// 	"time"
// )

// type ChildProcess struct{}

// func NewChildProcess() *ChildProcess {
// 	return &ChildProcess{}
// }

// // run executes the child process with proper namespace setup
// func (cp *ChildProcess) Run(args []string) error {
// 	rootfsPath := args[0]
// 	command := args[1:]
// 	if err := cp.signalParent(); err != nil {
// 		return fmt.Errorf("failed to signal parent: %w", err)
// 	}

// 	if err := cp.setupContainer(rootfsPath); err != nil {
// 		return fmt.Errorf("failed to setup container: %w", err)
// 	}

// 	return cp.execCommand(command)
// }

// // signalParent signals the parent process that namespace setup is complete
// func (cp *ChildProcess) signalParent() error {
// 	pipe := os.NewFile(3, "pipe")
// 	if pipe == nil {
// 		return fmt.Errorf("communication pipe not found")
// 	}
// 	defer pipe.Close()

// 	if _, err := pipe.Write([]byte{1}); err != nil {
// 		return fmt.Errorf("failed to write to pipe: %w", err)
// 	}

// 	return nil
// }

// // setupContainer performs container-specific setup
// func (cp *ChildProcess) setupContainer(rootfsPath string) error {
// 	// Generate unique hostname based on timestamp
// 	hostname := fmt.Sprintf("container-%d", time.Now().UnixNano())

// 	//1. Set hostname using the sethostname syscall
// 	if err := syscall.Sethostname([]byte(hostname)); err != nil {
// 		return fmt.Errorf("failed to set hostname: %w", err)
// 	}

// 	// 2. This ensures that mounts/unmounts in this new namespace don't affect the host.
// 	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
// 		return fmt.Errorf("failed to make root mount private: %w", err)
// 	}

// 	// 3. Prevent mount propagation
// 	// This ensures that mounts/unmounts in this new namespace don't affect the host.
// 	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
// 		return fmt.Errorf("failed to make root mount private: %w", err)
// 	}

// 	// 4. Change into the new rootfs directory.
// 	if err := os.Chdir(rootfsPath); err != nil {
// 		return fmt.Errorf("failed to chdir to rootfs: %w", err)
// 	}

// 	// 5. Create a directory for the old root and pivot into the new rootfs.
// 	// pivot_root requires the old root to be on a different mount from the new root.
// 	// By chdir'ing into the new root first, "." refers to the new root mount,
// 	// and ".old_root" is a directory inside it, satisfying the requirement.
// 	if err := os.MkdirAll(".old_root", 0700); err != nil {
// 		return fmt.Errorf("failed to create .old_root directory: %w", err)
// 	}

// 	if err := syscall.PivotRoot(".", ".old_root"); err != nil {
// 		return fmt.Errorf("failed to pivot_root: %w", err)
// 	}

// 	// 6. Change to the new root directory.
// 	if err := os.Chdir("/"); err != nil {
// 		return fmt.Errorf("failed to change directory: %w", err)
// 	}

// 	// 7. Unmount the old root. It's now at "/.old_root".
// 	// MNT_DETACH allows for a lazy unmount, which is safer.
// 	if err := syscall.Unmount("/.old_root", syscall.MNT_DETACH); err != nil {
// 		return fmt.Errorf("failed to unmount old_root: %w", err)
// 	}
// 	// Clean up the temporary directory.
// 	if err := os.RemoveAll("/.old_root"); err != nil {
// 		return fmt.Errorf("failed to remove old_root dir: %w", err)
// 	}

// 	// 8. Mount necessary virtual filesystems.
// 	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
// 		return fmt.Errorf("failed to mount /proc: %w", err)
// 	}
// 	if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
// 		return fmt.Errorf("failed to mount /sys: %w", err)
// 	}
// 	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "size=65536k"); err != nil {
// 		return fmt.Errorf("failed to mount /dev: %w", err)
// 	}

// 	os.Setenv("PS1", "container# ")
// 	os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

// 	return nil
// }

// func (cp *ChildProcess) execCommand(args []string) error {
// 	binary, err := exec.LookPath(args[0])
// 	if err != nil {
// 		return fmt.Errorf("executable not found: %s", args[0])
// 	}

// 	return syscall.Exec(binary, args, os.Environ())
// }

package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type ChildProcess struct{}

func NewChildProcess() *ChildProcess {
	return &ChildProcess{}
}

// run executes the child process with proper namespace and filesystem setup
func (cp *ChildProcess) Run(args []string) error {
	// The first arg is the rootfs path, the rest is the command.
	rootfsPath := args[0]
	command := args[1:]

	// Must setup the container *before* signaling the parent.
	// The parent waits for this signal to know the setup is complete.
	if err := cp.setupContainer(rootfsPath); err != nil {
		return fmt.Errorf("failed to setup container: %w", err)
	}

	if err := cp.signalParent(); err != nil {
		return fmt.Errorf("failed to signal parent: %w", err)
	}

	return cp.execCommand(command)
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

// setupContainer performs container-specific setup using pivot_root
func (cp *ChildProcess) setupContainer(rootfsPath string) error {
	// 1. Set a new hostname for the container
	if err := syscall.Sethostname([]byte("container")); err != nil {
		return fmt.Errorf("failed to set hostname: %w", err)
	}

	// 2. Prevent mount propagation
	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("failed to make root mount private: %w", err)
	}

	// 3. Bind mount the new rootfs to itself. This is a prerequisite for pivot_root.
	if err := syscall.Mount(rootfsPath, rootfsPath, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to bind mount rootfs: %w", err)
	}

	// 4. Change into the new rootfs directory.
	if err := os.Chdir(rootfsPath); err != nil {
		return fmt.Errorf("failed to chdir to rootfs: %w", err)
	}

	// 5. Create a directory for the old root and pivot into the new rootfs.
	// pivot_root requires the old root to be on a different mount from the new root.
	// By chdir'ing into the new root first, "." refers to the new root mount,
	// and ".old_root" is a directory inside it, satisfying the requirement.
	if err := os.MkdirAll(".old_root", 0700); err != nil {
		return fmt.Errorf("failed to create .old_root directory: %w", err)
	}

	if err := syscall.PivotRoot(".", ".old_root"); err != nil {
		return fmt.Errorf("failed to pivot_root: %w", err)
	}

	// 6. Change to the new root, which is now "/".
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("failed to chdir to new root: %w", err)
	}

	// 7. Unmount the old root. It's now at "/.old_root".
	if err := syscall.Unmount("/.old_root", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("failed to unmount old_root: %w", err)
	}
	if err := os.RemoveAll("/.old_root"); err != nil {
		return fmt.Errorf("failed to remove old_root dir: %w", err)
	}

	// 8. Mount necessary virtual filesystems.
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("failed to mount /proc: %w", err)
	}
	if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
		return fmt.Errorf("failed to mount /sys: %w", err)
	}
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "size=65536k"); err != nil {
		return fmt.Errorf("failed to mount /dev: %w", err)
	}

	// 9. Set up basic environment variables.
	os.Setenv("PS1", "container# ")
	os.Setenv("PATH", "/bin:/sbin:/usr/bin:/usr/sbin")

	return nil
}

func (cp *ChildProcess) execCommand(args []string) error {
	binary, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("executable not found: %s", args[0])
	}

	return syscall.Exec(binary, args, os.Environ())
}
