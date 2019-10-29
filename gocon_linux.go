package gocon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/docker/docker/pkg/mount"
	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

type Container struct {
	state
	PipeFD int `json:"pipe_fd"`
}

func (c *Container) Clone(args ...string) error {
	if err := c.createWorkspace(); err != nil {
		return fmt.Errorf("failed to create work dir: %s", err)
	}

	return c.clone(args...)
}

func (c *Container) createWorkspace() error {
	if err := createWorkDirIfNone(); err != nil {
		return err
	}
	if err := os.MkdirAll(c.workDir(), 0744); err != nil {
		return err
	}

	if err := c.createStateFile(); err != nil {
		return err
	}
	if err := c.createNamedPipe(); err != nil {
		return err
	}

	return nil
}

func createWorkDirIfNone() error {
	if _, err := os.Stat(workDir()); err == nil {
		return nil
	}

	return createWorkDir()
}

func createWorkDir() error {
	return os.MkdirAll(workDir(), 0744)
}

func (c *Container) createStateFile() error {
	name := c.stateFilename()
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return c.save()
}

func (c *Container) createNamedPipe() error {
	name := c.pipename()
	if err := unix.Mkfifo(name, 700); err != nil {
		return err
	}

	fd, err := unix.Open(name, os.O_RDONLY|unix.O_NONBLOCK, 700)
	if err != nil {
		return err
	}
	c.PipeFD = fd

	return nil
}

func (c *Container) clone(args ...string) error {
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

func (c *Container) Init(spec *specs.Spec) error {
	if err := c.load(); err != nil {
		return fmt.Errorf("failed to load: %s", err)
	}
	c.Version, c.Annotations = spec.Version, spec.Annotations
	c.Bundle, _ = filepath.Abs(spec.Root.Path)

	if spec.Hostname != "" {
		if err := unix.Sethostname([]byte(spec.Hostname)); err != nil {
			return fmt.Errorf("failed to set hostname")
		}
	}

	if err := c.mount(spec.Root, spec.Mounts); err != nil {
		return fmt.Errorf("failed to mount: %s", err)
	}

	if spec.Linux != nil {
		if err := c.limit(spec.Linux); err != nil {
			return fmt.Errorf("failed to limit: %s", err)
		}
	}

	if err := c.save(); err != nil {
		return fmt.Errorf("failed to save: %s", err)
	}

	if err := c.pivotRoot(spec.Root); err != nil {
		return fmt.Errorf("failed to pivot root: %s", err)
	}

	if err := c.exec(spec.Process); err != nil {
		return fmt.Errorf("failed to exec: %s", err)
	}

	return nil
}

func (c *Container) mount(root *specs.Root, ms []specs.Mount) error {
	var ops []string
	if root.Readonly {
		ops = append(ops, "ro")
	}

	for _, m := range ms {
		if err := mount.ForceMount(
			m.Source, filepath.Join(root.Path, m.Destination), m.Type, strings.Join(ops, ","),
		); err != nil {
			return err
		}
	}

	return nil
}

func (c *Container) limit(spec *specs.Linux) error {
	dir := cgroupDir(spec.CgroupsPath)
	if spec.Resources != nil && spec.Resources.CPU != nil {
		if err := c.limitCPU(dir, spec.Resources.CPU); err != nil {
			return err
		}
	}

	return nil
}

func cgroupDir(path string) string {
	dir := "/sys/fs/cgroup"
	if filepath.IsAbs(path) {
		return filepath.Join(dir, path)
	}

	return filepath.Join(dir, "gocon", path)
}

func (c *Container) limitCPU(dir string, spec *specs.LinuxCPU) error {
	dir = filepath.Join(dir, "cpu", "gocon")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if spec.Quota != nil {
		if err := ioutil.WriteFile(
			filepath.Join(dir, "cpu.cfs_quota_us"), []byte(fmt.Sprint(*spec.Quota)), 0755,
		); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(
		filepath.Join(dir, "tasks"), []byte(fmt.Sprint(os.Getpid())), 0755,
	); err != nil {
		return err
	}

	return nil
}

func (c *Container) pivotRoot(root *specs.Root) error {
	oldFs := "oldfs"
	if err := os.MkdirAll(filepath.Join(root.Path, oldFs), 0700); err != nil {
		return err
	}
	if err := unix.Mount(root.Path, root.Path, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return err
	}
	if err := unix.PivotRoot(root.Path, filepath.Join(root.Path, oldFs)); err != nil {
		return err
	}
	if err := unix.Chdir("/"); err != nil {
		return err
	}
	if err := unix.Unmount(oldFs, unix.MNT_DETACH); err != nil {
		return err
	}
	if err := os.RemoveAll(oldFs); err != nil {
		return err
	}

	return nil
}

func (c *Container) exec(proc *specs.Process) error {
	path, err := exec.LookPath(proc.Args[0])
	if err != nil {
		return err
	}

	return unix.Exec(path, proc.Args, os.Environ())
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

func (c *Container) pipename() string {
	return filepath.Join(c.workDir(), "pipe.fifo")
}

func (c *Container) workDir() string {
	return filepath.Join(workDir(), c.ID)
}

func workDir() string {
	return filepath.Join("/run", "gocon")
}
