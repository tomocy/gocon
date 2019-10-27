package client

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

func New() Client {
	return newClient()
}

type Client interface {
	Run(args []string) error
}

func newClient() *client {
	c := new(client)
	c.setUp()

	return c
}

type client struct {
	cli.App
	container func(string) Container
}

func (c *client) setUp() {
	c.App = *cli.NewApp()
	c.setBasic()
	c.setCommands()
}

func (c *client) setBasic() {
	c.Name = "gocon"
	c.Usage = "a CLI client which implements OCI runtime specification and is presented in Go Confenrece'19 Autumn in Tokyo"
}

func (c *client) setCommands() {
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

func (c *client) state(ctx *cli.Context) error {
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

func (c *client) create(ctx *cli.Context) error {
	id, specPath := ctx.Args().First(), ctx.Args().Get(1)

	return c.container(id).Clone("init", id, specPath)
}

func (c *client) init(ctx *cli.Context) error {
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

func (c *client) init(ctx *cli.Context) error {
	return errors.New("not implemented")
}

func (c *client) start(ctx *cli.Context) error {
	return errors.New("not implemented")
}

func (c *client) kill(ctx *cli.Context) error {
	return errors.New("not implemented")
}

func (c *client) delete(ctx *cli.Context) error {
	return errors.New("not implemented")
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
