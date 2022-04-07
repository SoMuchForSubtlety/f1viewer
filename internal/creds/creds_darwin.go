//go:build darwin
// +build darwin

package creds

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "f1viewer"
	userKey     = "username"
	passKey     = "password"
	tokenKey    = "token"
)

func LoadCredentials() (string, string, string, error) {
	username, err := keyring.Get(serviceName, userKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get username: %w", err)
	}
	password, err := keyring.Get(serviceName, passKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get password: %w", err)
	}
	token, err := keyring.Get(serviceName, tokenKey)
	if err != nil {
		return username, password, "", nil
	}
	return username, password, token, nil
}

func SaveCredentials(username, password, token string) error {
	err := keyring.Set(serviceName, userKey, username)
	if err != nil {
		return fmt.Errorf("failed to save username: %w", err)
	}
	keyring.Set(serviceName, passKey, password)
	if err != nil {
		return fmt.Errorf("failed to save password: %w", err)
	}
	keyring.Set(serviceName, tokenKey, token)
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	return nil
}

func RemoveCredentials() error {
	if err := keyring.Delete(serviceName, userKey); err != nil {
		return fmt.Errorf("failed to delete username: %w", err)
	}
	if err := keyring.Delete(serviceName, passKey); err != nil {
		return fmt.Errorf("failed to delete password: %w", err)
	}
	if err := keyring.Delete(serviceName, tokenKey); err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}
