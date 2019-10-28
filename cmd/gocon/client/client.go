package client

import "github.com/urfave/cli"

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
