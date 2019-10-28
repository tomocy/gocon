package gocon

import (
	"os"
	"path/filepath"
)

func (c *Container) specFilename() string {
	return filepath.Join(workDir(), c.ID, "spec.json")
}

func createWorkDirIfNone() error {
	dir := workDir()
	if _, err := os.Stat(dir); err == nil {
		return nil
	}

	return createWorkDir()
}

func createWorkDir() error {
	return os.MkdirAll(workDir(), 0744)
}

func workDir() string {
	return filepath.Join("/run", "gocon")
}
