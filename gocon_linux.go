package gocon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (c *Container) save() error {
	dst, err := os.OpenFile(c.stateFilename(), os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return fmt.Errorf("failed to open spec file: %s", err)
	}
	defer dst.Close()

	return json.NewEncoder(dst).Encode(c)
}

func (c *Container) load() error {
	src, err := os.Open(c.stateFilename())
	if err != nil {
		return fmt.Errorf("failed to open spec file: %s", err)
	}
	defer src.Close()

	return json.NewDecoder(src).Decode(c)
}

func (c *Container) stateFilename() string {
	return filepath.Join(c.workDir(), "state.json")
}

func (c *Container) workDir() string {
	return filepath.Join(workDir(), c.ID)
}

func workDir() string {
	return filepath.Join("/run", "gocon")
}
