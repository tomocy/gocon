package gocon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

func (c *Container) Clone(args ...string) error {
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &unix.SysProcAttr{
		Cloneflags: unix.CLONE_NEWIPC | unix.CLONE_NEWNET | unix.CLONE_NEWNS | unix.CLONE_NEWPID | unix.CLONE_NEWUSER | unix.CLONE_NEWUTS,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	return cmd.Run()
}

func (c *Container) Init(spec *specs.Spec) error {
	if err := unix.Sethostname([]byte(c.ID)); err != nil {
		return fmt.Errorf("failed to set hostname: %s", err)
	}
	if err := c.mount(spec.Root, spec.Mounts); err != nil {
		return fmt.Errorf("failed to mount: %s", err)
	}

	return nil
}

func (c *Container) mount(root *specs.Root, ms []specs.Mount) error {
	var flags uintptr
	if root.Readonly {
		flags = unix.MS_RDONLY
	}

	ms = append(defaultFs, ms...)
	for _, m := range ms {
		dst := filepath.Join(root.Path, m.Destination)
		if err := os.MkdirAll(dst, 0755); err != nil {
			return err
		}

		flags |= unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV
		if err := unix.Mount(m.Source, dst, m.Type, flags, ""); err != nil {
			return err
		}
	}

	return nil
}

var defaultFs = []specs.Mount{
	{Destination: "/proc", Type: "proc", Source: "/proc"},
}
