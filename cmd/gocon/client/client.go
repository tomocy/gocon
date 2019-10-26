package client

import "github.com/urfave/cli"

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
