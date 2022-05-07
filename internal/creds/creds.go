//go:build !darwin
// +build !darwin

package creds

import (
	"fmt"
	"path"

	"github.com/99designs/keyring"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/config"
)

const serviceName = "f1viewer"

func LoadCredentials() (string, string, string, error) {
	ring, err := openRing()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to open secret store: %w", err)
	}

	username, err := ring.Get("username")
	if err != nil {
		return "", "", "", fmt.Errorf("Could not get username: %w", err)
	}

	password, err := ring.Get("password")
	if err != nil {
		return "", "", "", fmt.Errorf("Could not get password: %w", err)
	}

	token, err := ring.Get("token")
	if err != nil {
		return string(username.Data), string(password.Data), "", nil
	}
	return string(username.Data), string(password.Data), string(token.Data), nil
}

func SaveCredentials(username, password, token string) error {
	conf, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	ring, err := openRing()
	if err != nil {
		return fmt.Errorf("failed to open secret store: %w", err)
	}

	uKey := "username"
	pKey := "password"

	pfx := conf.KeyRingPrefix
	if pfx != "" {
		uKey = path.Join(pfx, uKey)
		pKey = path.Join(pfx, pKey)
	}

	err = ring.Set(keyring.Item{
		Description: "F1TV username",
		Key:         uKey,
		Data:        []byte(username),
	})
	if err != nil {
		return fmt.Errorf("could not save username %w", err)
	}

	err = ring.Set(keyring.Item{
		Description: "F1TV password",
		Key:         pKey,
		Data:        []byte(password),
	})
	if err != nil {
		return fmt.Errorf("could not save password %w", err)
	}

	err = ring.Set(keyring.Item{
		Description: "F1TV subscription token",
		Key:         "token",
		Data:        []byte(token),
	})
	if err != nil {
		return fmt.Errorf("could not save token %w", err)
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
