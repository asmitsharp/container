package container

import "github.com/ashmitsharp/container/internal/namespaces"

type ContainerConfig struct {
	Image      string
	Command    []string
	Namespaces namespaces.NamespaceFlags
	Hostname   string
	WorkingDir string
	Env        []string
	Rootfs     string
}

func DefaultContainerConfig() *ContainerConfig {
	return &ContainerConfig{
		Image:      "default",
		Command:    []string{"/bin/sh"},
		Namespaces: *namespaces.DefaultNamespace(),
		Hostname:   "container-host",
		WorkingDir: "/",
		Env:        []string{},
		Rootfs:     "./busybox-rootfs",
	}
}
