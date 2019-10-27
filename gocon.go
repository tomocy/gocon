package gocon

import (
	"errors"
	"os"

	"github.com/opencontainers/runtime-spec/specs-go"
)

func New(id string) *Container {
	return &Container{
		state: state(specs.State{
			ID: id,
		}),
	}
}

type Container struct {
	state
}

type state specs.State

func (c *Container) State() (*specs.State, error) {
	return nil, errors.New("not implemented")
}

func (c *Container) Start() error {
	return errors.New("not implemented")
}

func (c *Container) Kill(os.Signal) error {
	return errors.New("not implemented")
}

func (c *Container) Delete() error {
	return errors.New("not implemented")
}
