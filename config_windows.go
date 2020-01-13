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
	return filepath.Dir(ex) + "\\", nil
}

func getLogPath(cfg config) (string, error) {
	if cfg.LogLocation != "" {
		path := filepath.Dir(cfg.LogLocation)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				return path, err
			}
		}
	}
	return getConfigPath()
}
