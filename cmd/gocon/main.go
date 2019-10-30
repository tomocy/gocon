package main

import (
	"fmt"
	"os"

	"github.com/tomocy/gocon"
	"github.com/tomocy/gocon/cmd/gocon/client"
)

func main() {
	c := client.New(func(id string) client.Container {
		return gocon.New(id)
	})
	if err := c.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
