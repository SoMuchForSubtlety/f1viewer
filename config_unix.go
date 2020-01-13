// +build !windows

package main

import (
	"os"
)

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := home + "/.config/F1viewer/"
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return path, err
		}
	}
	return path, nil
}

func getLogPath(cfg config) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := home + "/.local/share/F1viewer/"
	if cfg.LogLocation != "" {
		path = cfg.LogLocation
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return path, err
		}
	}
	return path, nil
}
