package gocon

import "github.com/opencontainers/runtime-spec/specs-go"

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
