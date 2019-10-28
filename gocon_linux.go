package gocon

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

func (c *Container) Clone(args ...string) error {
	cmd := c.buildCloneCmd(args...)
	if err := cmd.Start(); err != nil {
		return err
	}

	c.Pid = cmd.Process.Pid
	if err := c.save(); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to save: %s", err)
	}

	return cmd.Wait()
}

func (c *Container) buildCloneCmd(args ...string) *exec.Cmd {
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &unix.SysProcAttr{
		Cloneflags: unix.CLONE_NEWIPC | unix.CLONE_NEWNET | unix.CLONE_NEWNS |
			unix.CLONE_NEWPID | unix.CLONE_NEWUSER | unix.CLONE_NEWUTS,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	return cmd
}

func (c *Container) save() error {
	dst, err := os.OpenFile(c.stateFilename(), os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return fmt.Errorf("failed to open spec file: %s", err)
	}
	defer dst.Close()

	return json.NewEncoder(dst).Encode(c)
}

func (c *Container) load() error {
	src, err := os.Open(c.stateFilename())
	if err != nil {
		return fmt.Errorf("failed to open spec file: %s", err)
	}
	defer src.Close()

	return json.NewDecoder(src).Decode(c)
}

func (c *Container) stateFilename() string {
	return filepath.Join(c.workDir(), "state.json")
}

func (c *Container) workDir() string {
	return filepath.Join(workDir(), c.ID)
}

func workDir() string {
	return filepath.Join("/run", "gocon")
}
