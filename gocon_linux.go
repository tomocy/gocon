package gocon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

func (c *Container) Clone(args ...string) error {
	if err := c.createWorkspace(); err != nil {
		return fmt.Errorf("failed to create workspace: %s", err)
	}

	cmd := c.buildCloneCmd(args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	c.Pid = cmd.Process.Pid
	c.Status = statusCreating

	if err := c.save(); err != nil {
		return fmt.Errorf("failed to save: %s", err)
	}

	return cmd.Wait()
}

func (c *Container) createWorkspace() error {
	dir := c.workspace()
	if err := os.MkdirAll(dir, 0744); err != nil {
		return err
	}

	if err := c.createStateFile(); err != nil {
		return err
	}

	return c.createPipe()
}

func (c *Container) createStateFile() error {
	name := c.stateFilename()
	dst, err := os.Create(name)
	if err != nil {
		return err
	}
	defer dst.Close()

	return json.NewEncoder(dst).Encode(specs.State{})
}

func (c *Container) createPipe() error {
	name := c.pipeFilename()

	return unix.Mkfifo(name, 0744)
}

func (c *Container) buildCloneCmd(args ...string) *exec.Cmd {
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

	return cmd
}

func (c *Container) Init(spec *specs.Spec) error {
	if err := unix.Sethostname([]byte(c.ID)); err != nil {
		return fmt.Errorf("failed to set hostname: %s", err)
	}
	if err := c.mount(spec.Root, spec.Mounts); err != nil {
		return fmt.Errorf("failed to mount: %s", err)
	}
	if err := c.enable(spec.Root, bins, libs...); err != nil {
		return fmt.Errorf("failed to enable: %s", err)
	}
	if spec.Linux != nil {
		if err := c.limit(spec.Linux); err != nil {
			return fmt.Errorf("failed to limit: %s", err)
		}
	}
	if err := c.pivotRoot(spec.Root); err != nil {
		return fmt.Errorf("failed to pivot root: %s", err)
	}
	if err := c.exec(spec.Process); err != nil {
		return fmt.Errorf("failed to exec: %s", err)
	}

	return nil
}

func (c *Container) load() error {
	name := c.stateFilename()
	src, err := os.Open(name)
	if err != nil {
		return err
	}
	defer src.Close()

	return json.NewDecoder(src).Decode(&c.state)
}

func (c *Container) meet(spec *specs.Spec) {
	c.Version = spec.Version
	c.Bundle = spec.Root.Path
	c.Annotations = spec.Annotations
}

func (c *Container) save() error {
	name := c.stateFilename()
	dst, err := os.OpenFile(name, os.O_WRONLY, 0744)
	if err != nil {
		return err
	}
	defer dst.Close()

	return json.NewEncoder(dst).Encode(c.state)
}

func (c *Container) stateFilename() string {
	return filepath.Join(c.workspace(), "spec.json")
}

func (c *Container) pipeFilename() string {
	return filepath.Join(c.workspace(), "pipe.fifo")
}

func (c *Container) workspace() string {
	return filepath.Join(workSpacesDir, c.ID)
}

const workSpacesDir = "/run/gocon"

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

func (c *Container) enable(root *specs.Root, bins []string, libs ...string) error {
	if 1 <= len(libs) {
		if err := c.ensure(root, libs); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Join(root.Path, "bin"), 0755); err != nil {
		return err
	}

	for _, bin := range bins {
		if err := copyFile(bin, filepath.Join(root.Path, bin)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Container) ensure(root *specs.Root, libs []string) error {
	if err := os.MkdirAll(filepath.Join(root.Path, "lib"), 0755); err != nil {
		return err
	}

	for _, lib := range libs {
		if err := copyFile(lib, filepath.Join(root.Path, lib)); err != nil {
			return err
		}
	}

	return nil
}

var (
	bins = []string{"/bin/sh", "/bin/ls", "/bin/ps", "/bin/cat", "/bin/date", "/bin/echo"}
	libs = []string{"/lib/ld-musl-x86_64.so.1"}
)

func copyFile(src, dst string) error {
	read, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, read, 0755)
}

func (c *Container) limit(spec *specs.Linux) error {
	if spec.Resources != nil && spec.Resources.CPU != nil {
		if err := c.limitCPU(spec.CgroupsPath, spec.Resources.CPU); err != nil {
			return err
		}
	}

	return nil
}

func (c *Container) limitCPU(path string, spec *specs.LinuxCPU) error {
	dir := c.cgroupsDir("cpu", path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if spec.Shares != nil {
		if err := c.limitCPUShares(dir, *spec.Shares); err != nil {
			return err
		}
	}
	if spec.Quota != nil {
		if err := c.limitCPUQuota(dir, *spec.Quota); err != nil {
			return err
		}
	}
	if spec.Period != nil {
		if err := c.limitCPUPeriod(dir, *spec.Period); err != nil {
			return err
		}
	}
	if spec.RealtimeRuntime != nil {
		if err := c.limitCPURealtimeRuntime(dir, *spec.RealtimeRuntime); err != nil {
			return err
		}
	}
	if spec.RealtimePeriod != nil {
		if err := c.limitCPURealtimePeriod(dir, *spec.RealtimePeriod); err != nil {
			return err
		}
	}

	return c.addLimit(filepath.Join(dir, "tasks"), fmt.Sprint(os.Getpid()))
}

func (c *Container) limitCPUShares(dir string, shares uint64) error {
	return c.addLimit(filepath.Join(dir, "cpu.shares"), fmt.Sprint(shares))
}

func (c *Container) limitCPUQuota(dir string, quota int64) error {
	return c.addLimit(filepath.Join(dir, "cpu.cfs_quota_us"), fmt.Sprint(quota))
}

func (c *Container) limitCPUPeriod(dir string, period uint64) error {
	return c.addLimit(filepath.Join(dir, "cpu.cfs_period_us"), fmt.Sprint(period))
}

func (c *Container) limitCPURealtimeRuntime(dir string, runtime int64) error {
	log.Println(runtime)
	return c.addLimit(filepath.Join(dir, "cpu.rt_runtime_us"), fmt.Sprint(runtime))
}

func (c *Container) limitCPURealtimePeriod(dir string, period uint64) error {
	return c.addLimit(filepath.Join(dir, "cpu.rt_period_us"), fmt.Sprint(period))
}

func (c *Container) addLimit(name, limit string) error {
	return ioutil.WriteFile(name, []byte(limit), 644)
}

func (c *Container) cgroupsDir(kind, path string) string {
	if path == "" {
		return filepath.Join(cgroupsDir, kind, "gocon", c.ID)
	}
	if !filepath.IsAbs(path) {
		return filepath.Join(cgroupsDir, kind, "gocon", c.ID, path)
	}

	return filepath.Join(cgroupsDir, kind, path)
}

const cgroupsDir = "/sys/fs/cgroup"

func (c *Container) pivotRoot(root *specs.Root) error {
	oldFs := "oldfs"
	if err := os.MkdirAll(filepath.Join(root.Path, oldFs), 0755); err != nil {
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
	cmd := exec.Command(proc.Args[0], proc.Args[1:]...)
	cmd.Env = proc.Env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	return cmd.Run()
}

const (
	statusCreating = "creating"
)
