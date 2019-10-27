package client

import (
	"encoding/json"
	"errors"
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
	c.Usage = "a CLI client which implements OCI runtime specification and is presented in Go Confenrece'19 Autumn in Tokyo"
}

func (c *Client) setCommands() {
	c.Commands = []cli.Command{
		cli.Command{
			Name:      "state",
			Usage:     "state the container of the given ID",
			ArgsUsage: "id",
			Action:    c.state,
		},
		cli.Command{
			Name:      "create",
			Usage:     "create a container with the give ID and the given path to bundle",
			ArgsUsage: "id path-to-bundle",
			Action:    c.create,
		},
		cli.Command{
			Name:   "init",
			Action: c.init,
			Hidden: true,
		},
		cli.Command{
			Name:      "start",
			Usage:     "start a command the container of the given ID is waiting to exec",
			ArgsUsage: "id",
			Action:    c.start,
		},
		cli.Command{
			Name:      "kill",
			Usage:     "kill the container of the given ID with the given signal",
			ArgsUsage: "id signal(default: SIGTERM)",
			Action:    c.kill,
		},
		cli.Command{
			Name:      "delete",
			Usage:     "delete the kibidango of the given ID",
			ArgsUsage: "id",
			Action:    c.delete,
		},
	}
}

func (c *Client) state(ctx *cli.Context) error {
	id := ctx.Args().First()
	if err := validateID(id); err != nil {
		return err
	}

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
	if err := validateID(id); err != nil {
		return err
	}

	spec, err := loadSpec(specPath)
	if err != nil {
		return err
	}

	return c.container(id).Init(spec)
}

func loadSpec(name string) (*specs.Spec, error) {
	src, err := os.Open(name)
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
	if err := validateID(id); err != nil {
		return err
	}

	return c.container(id).Start()
}

func (c *Client) kill(ctx *cli.Context) error {
	id, sigName := ctx.Args().First(), ctx.Args().Get(1)
	if err := validateID(id); err != nil {
		return err
	}

	sig, err := mapSignal(sigName)
	if err != nil {
		return err
	}

	return c.container(id).Kill(sig)
}

func mapSignal(name string) (os.Signal, error) {
	if n, err := strconv.Atoi(name); err == nil {
		return syscall.Signal(n), nil
	}

	trimed := strings.TrimLeft(strings.ToUpper(name), "SIG")
	if signal, ok := sigMap[trimed]; ok {
		return signal, nil
	}

	return nil, fmt.Errorf("no such signal: %s", name)
}

var sigMap = map[string]os.Signal{
	"HUP": syscall.SIGHUP, "INT": syscall.SIGINT, "QUIT": syscall.SIGQUIT, "ILL": syscall.SIGILL,
	"TRAP": syscall.SIGTRAP, "ABRT": syscall.SIGABRT, "FPE": syscall.SIGFPE, "KILL": syscall.SIGKILL,
	"EGV": syscall.SIGSEGV, "PIPE": syscall.SIGPIPE, "ALRM": syscall.SIGALRM, "TERM": syscall.SIGTERM,
}

func (c *Client) delete(ctx *cli.Context) error {
	id := ctx.Args().First()
	if err := validateID(id); err != nil {
		return err
	}

	return c.container(id).Delete()
}

func validateID(id string) error {
	if id == "" {
		return errors.New("id should not be empty")
	}

	return nil
}

type Container interface {
	State() (*specs.State, error)
	Clone(args ...string) error
	Init(*specs.Spec) error
	Start() error
	Kill(os.Signal) error
	Delete() error
}
