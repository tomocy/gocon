package gocon

import (
	"os"
	"path/filepath"
)

func createWorkDir() error {
	return os.MkdirAll(workDir(), 0744)
}

func workDir() string {
	return filepath.Join("/run", "gocon")
}
