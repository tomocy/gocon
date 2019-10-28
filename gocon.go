package gocon

import "github.com/opencontainers/runtime-spec/specs-go"

type Container struct {
	state
}

type state specs.State
