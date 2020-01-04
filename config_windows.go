package main

import (
	"os"
	"path/filepath"
)

func getConfigPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}

func getLogPath() (string, error) {
	return getConfigPath()
}
