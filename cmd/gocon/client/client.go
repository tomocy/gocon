package client

import (
	"encoding/json"
	"os"

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
}

func (c *Client) setBasic() {
	c.Name = "gocon"
	c.Usage = "a container runtime which implements OCU runtime specification"
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

type Container interface {
	State() (*specs.State, error)
	Clone(...string) error
	Init(*specs.Spec) error
	Start() error
	Kill(os.Signal) error
	Delete() error
}
