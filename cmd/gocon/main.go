package main

import (
	"fmt"
	"os"

	"github.com/tomocy/gocon/cmd/gocon/client"
)

func main() {
	c := client.New(nil)
	if err := c.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
