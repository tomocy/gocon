package gocon

import "github.com/opencontainers/runtime-spec/specs-go"

func New(id string) *Container {
	return &Container{
		state: state{
			ID: id,
		},
	}
}

type state specs.State
