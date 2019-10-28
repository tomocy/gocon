package client

import (
	"os"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

func New() *Client {
	c := new(Client)
	c.setUp()

	return c
}

type Client struct {
	cli.App
}

func (c *Client) setUp() {
	c.App = *cli.NewApp()
	c.setBasic()
}

func (c *Client) setBasic() {
	c.Name = "gocon"
	c.Usage = "a container runtime which implements OCU runtime specification"
}

type Container interface {
	State() (*specs.State, error)
	Clone(...string) error
	Init(*specs.Spec) error
	Start() error
	Kill(os.Signal) error
	Delete() error
}
