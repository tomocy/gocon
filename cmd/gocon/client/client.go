package client

import (
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
}

func (c *client) setBasic() {
	c.Name = "gocon"
	c.Usage = "a CLI client which implements OCI runtime specification and is presented in Go Confenrece'19 Autumn in Tokyo"
}

type Container interface {
	State() (*specs.State, error)
	Clone(args ...string) error
	Init(*specs.Spec) error
	Start() error
	Kill(os.Signal) error
	Delete() error
}
