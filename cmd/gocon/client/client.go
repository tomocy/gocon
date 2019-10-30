package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

func New(fn func(string) Container) *Client {
	c := &Client{
		container: fn,
	}
	c.setUp()

	return c
}

type Client struct {
	cli.App
	container func(string) Container
}

func (c *Client) setUp() {
	c.App = *cli.NewApp()
	c.setBasic()
	c.setCommands()
}

func (c *Client) setBasic() {
	c.Name = "gocon"
	c.Usage = "a container runtime which implements OCI runtime specification"
}

func (c *Client) setCommands() {
	c.Commands = []cli.Command{
		{
			Name:      "state",
			Usage:     "state the container of the given id",
			ArgsUsage: "id",
			Action:    c.state,
		},
		{
			Name:      "create",
			Usage:     "create a container of the given id",
			ArgsUsage: "id spec-path",
			Action:    c.create,
		},
		{
			Name:   "init",
			Action: c.init,
			Hidden: true,
		},
		{
			Name:      "start",
			Usage:     "start the container of the given id",
			ArgsUsage: "id",
			Action:    c.start,
		},
		{
			Name:      "kill",
			Usage:     "send the given signal to the container of the given id",
			ArgsUsage: "id signal",
			Action:    c.kill,
		},
		{
			Name:      "delete",
			Usage:     "delete the container of the given id",
			ArgsUsage: "id",
			Action:    c.delete,
		},
	}
}

func (c *Client) state(ctx *cli.Context) error {
	id := ctx.Args().First()
	state, err := c.container(id).State()
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(state)
}

func (c *Client) create(ctx *cli.Context) error {
	id, specPath := ctx.Args().First(), ctx.Args().Get(1)

	return c.container(id).Clone("init", id, specPath)
}

func (c *Client) init(ctx *cli.Context) error {
	id, specPath := ctx.Args().First(), ctx.Args().Get(1)
	spec, err := loadSpec(specPath)
	if err != nil {
		return err
	}

	return c.container(id).Init(spec)
}

func loadSpec(path string) (*specs.Spec, error) {
	src, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	spec := new(specs.Spec)
	if err := json.NewDecoder(src).Decode(spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func (c *Client) start(ctx *cli.Context) error {
	id := ctx.Args().First()

	return c.container(id).Start()
}

func (c *Client) kill(ctx *cli.Context) error {
	id, sigStr := ctx.Args().First(), ctx.Args().Get(1)
	sig, err := mapSignal(sigStr)
	if err != nil {
		return err
	}

	return c.container(id).Kill(sig)
}

func mapSignal(target string) (os.Signal, error) {
	if n, err := strconv.Atoi(target); err == nil {
		return syscall.Signal(n), nil
	}

	trimed := strings.TrimLeft(strings.ToUpper(target), "SIG")
	if signal, ok := sigMap[trimed]; ok {
		return signal, nil
	}

	return nil, fmt.Errorf("no such signal")
}

var sigMap = map[string]os.Signal{
	"HUP": syscall.SIGHUP, "INT": syscall.SIGINT, "QUITA": syscall.SIGQUIT,
	"ILL": syscall.SIGILL, "TRAP": syscall.SIGTRAP, "ABRT": syscall.SIGABRT,
	"FPE": syscall.SIGFPE, "KILL": syscall.SIGKILL, "SEGV": syscall.SIGSEGV,
	"PIPE": syscall.SIGPIPE, "ALRM": syscall.SIGALRM, "TEM": syscall.SIGTERM,
}

func (c *Client) delete(ctx *cli.Context) error {
	id := ctx.Args().First()

	return c.container(id).Delete()
}

type Container interface {
	State() (*specs.State, error)
	Clone(...string) error
	Init(*specs.Spec) error
	Start() error
	Kill(os.Signal) error
	Delete() error
}
