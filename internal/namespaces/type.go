//go:build linux

package namespaces

import "syscall"

type NamespaceFlags struct {
	UTS   bool `json:"uts"`
	PID   bool `json:"pid"`
	NET   bool `json:"net"`
	MOUNT bool `json:"mount"`
	IPC   bool `json:"ipc"`
	USER  bool `json:"user"`
}

func DefaultNamespace() *NamespaceFlags {
	return &NamespaceFlags{
		UTS:   true,
		PID:   true,
		NET:   false, // to be implemented
		MOUNT: true,
		IPC:   true,
		USER:  false, // to be implemented
	}
}

func (nf *NamespaceFlags) ToCloneFlags() uintptr {
	var flags uintptr
	if nf.UTS {
		flags |= syscall.CLONE_NEWUTS
	}
	if nf.PID {
		flags |= syscall.CLONE_NEWPID
	}
	if nf.IPC {
		flags |= syscall.CLONE_NEWIPC
	}
	if nf.MOUNT {
		flags |= syscall.CLONE_NEWNS
	}
	if nf.NET {
		flags |= syscall.CLONE_NEWNET
	}
	if nf.USER {
		flags |= syscall.CLONE_NEWUSER
	}

	return flags
}
