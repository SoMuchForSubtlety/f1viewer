package secret

import (
	"errors"
	"fmt"

	"github.com/SoMuchForSubtlety/keyring"
)

type SecretStore struct {
	ring keyring.Keyring
}

func (s *SecretStore) LoadCredentials() (string, string, string, error) {
	if s.ring == nil {
		err := s.openRing()
		if err != nil {
			return "", "", "", err
		}
	}

	if s.ring == nil {
		return "", "", "", errors.New("No keyring configured")
	}
	username, err := s.ring.Get("username")
	if err != nil {
		return "", "", "", fmt.Errorf("Could not get username: %w", err)
	}

	password, err := s.ring.Get("password")
	if err != nil {
		return "", "", "", fmt.Errorf("Could not get password: %w", err)
	}
	token, err := s.ring.Get("skylarkToken")
	if err != nil {
		return "", "", "", fmt.Errorf("Could not get auth token: %w", err)
	}
	return string(username.Data), string(password.Data), string(token.Data), nil
}

func (s *SecretStore) SaveCredentials(username, password, authtoken string) error {
	if s.ring == nil {
		err := s.openRing()
		if err != nil {
			return err
		}
	}

	if s.ring == nil {
		return errors.New("No keyring configured")
	}
	err := s.ring.Set(keyring.Item{
		Description: "F1TV username",
		Key:         "username",
		Data:        []byte(username),
	})
	if err != nil {
		return fmt.Errorf("[ERROR] could not save username %w", err)
	}

	err = s.ring.Set(keyring.Item{
		Description: "F1TV password",
		Key:         "password",
		Data:        []byte(password),
	})
	if err != nil {
		return fmt.Errorf("[ERROR] could not save password %w", err)
	}

	err = s.ring.Set(keyring.Item{
		Description: "F1TV auth token",
		Key:         "skylarkToken",
		Data:        []byte(authtoken),
	})
	if err != nil {
		return fmt.Errorf("[ERROR] could not save auth token %w", err)
	}
	return nil
}

func (s *SecretStore) RemoveCredentials() error {
	if s.ring == nil {
		err := s.openRing()
		if err != nil {
			return err
		}
	}

	if s.ring == nil {
		return errors.New("No keyring configured")
	}
	err := s.ring.Remove("username")
	if err != nil {
		return fmt.Errorf("Could not remove username: %w", err)
	}

	err = s.ring.Remove("password")
	if err != nil {
		return fmt.Errorf("Could not remove password: %w", err)
	}

	err = s.ring.Remove("skylarkToken")
	if err != nil {
		return fmt.Errorf("Could not remove auth token: %w", err)
	}
	return nil
}

func (s *SecretStore) openRing() error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "f1viewer",
		AllowedBackends: []keyring.BackendType{
			keyring.KWalletBackend,
			keyring.PassBackend,
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
		},
	})
	if err != nil {
		return err
	}
	s.ring = ring
	return nil
}
