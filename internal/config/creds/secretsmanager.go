package creds

import (
	"fmt"

	"github.com/danhale-git/runrdp/internal/config/creds/secretsmanager"
)

// SecretsManagerStruct a struct of type creds.SecretsManager.
func SecretsManagerStruct() interface{} {
	return &SecretsManager{}
}

// Validate returns an error if a config field is invalid.
func (s *SecretsManager) Validate() error {
	// Update test when implemented
	return nil
}

// SecretsManager implements Cred and retrieves a username and password from AWS Secrets Manager.
type SecretsManager struct {
	UsernameID string
	PasswordID string
	Profile    string
	Region     string
}

// Retrieve returns the values for the configured Secrets Manager keys.
func (s *SecretsManager) Retrieve() (string, string, error) {
	svc := secretsmanager.NewSession(s.Profile, s.Region)

	username, err := secretsmanager.Get(svc, s.UsernameID)
	if err != nil {
		return "", "", fmt.Errorf("retrieving username: %s", err)
	}

	password, err := secretsmanager.Get(svc, s.PasswordID)
	if err != nil {
		return "", "", fmt.Errorf("retrieving password: %s", err)
	}

	return username, password, nil
}
