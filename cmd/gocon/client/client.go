package client

import (
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
	return errors.New("not implemented")
}

func (c *client) create(ctx *cli.Context) error {
	return errors.New("not implemented")
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

type Container interface {
	State() (*specs.State, error)
	Clone(args ...string) error
	Init(*specs.Spec) error
	Start() error
	Kill(os.Signal) error
	Delete() error
}
