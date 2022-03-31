package creds

import (
	"errors"
	"os"
)

func LoadEnvCredentials() (string, string, error) {
	username, password := os.Getenv("F1VIEWER_USERNAME"), os.Getenv("F1VIEWER_PASSWORD")

	if username == "" || password == "" {
		return "", "", errors.New("missing F1VIEWER_USERNAME or F1VIEWER_PASSWORD")
	}

	return username, password, nil
}
