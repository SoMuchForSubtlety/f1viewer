//go:build !darwin
// +build !darwin

package creds

import (
	"fmt"

	"github.com/99designs/keyring"
)

const serviceName = "f1viewer"

func LoadCredentials() (string, string, error) {
	ring, err := openRing()
	if err != nil {
		return "", "", fmt.Errorf("failed to open secret store: %w", err)
	}

	username, err := ring.Get("username")
	if err != nil {
		return "", "", fmt.Errorf("Could not get username: %w", err)
	}

	password, err := ring.Get("password")
	if err != nil {
		return "", "", fmt.Errorf("Could not get password: %w", err)
	}
	return string(username.Data), string(password.Data), nil
}

func SaveCredentials(username, password string) error {
	ring, err := openRing()
	if err != nil {
		return fmt.Errorf("failed to open secret store: %w", err)
	}

	err = ring.Set(keyring.Item{
		Description: "F1TV username",
		Key:         "username",
		Data:        []byte(username),
	})
	if err != nil {
		return fmt.Errorf("could not save username %w", err)
	}

	err = ring.Set(keyring.Item{
		Description: "F1TV password",
		Key:         "password",
		Data:        []byte(password),
	})
	if err != nil {
		return fmt.Errorf("could not save password %w", err)
	}
	return nil
}

func RemoveCredentials() error {
	ring, err := openRing()
	if err != nil {
		return fmt.Errorf("failed to open secret store: %w", err)
	}

	err = ring.Remove("username")
	if err != nil {
		return fmt.Errorf("Could not remove username: %w", err)
	}

	err = ring.Remove("password")
	if err != nil {
		return fmt.Errorf("Could not remove password: %w", err)
	}

	return nil
}

func openRing() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: serviceName,
		AllowedBackends: []keyring.BackendType{
			keyring.PassBackend,
			keyring.SecretServiceBackend,
			keyring.KWalletBackend,
			keyring.WinCredBackend,
		},
	})
}
