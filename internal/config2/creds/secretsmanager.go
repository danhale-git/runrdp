package creds

import (
	"fmt"

	"github.com/danhale-git/runrdp/internal/aws/secrets"
)

// SecretsManagerCred implements Cred and retrieves a username and password from AWS Secrets Manager.
type SecretsManagerCred struct {
	UsernameID string
	PasswordID string
	Profile    string
	Region     string
}

// Retrieve returns the values for the configured Secrets Manager keys.
func (s *SecretsManagerCred) Retrieve() (string, string, error) {
	username, err := secrets.Get(s.Profile, s.Region, s.UsernameID)
	if err != nil {
		return "", "", fmt.Errorf("retrieving username: %s", err)
	}

	password, err := secrets.Get(s.Profile, s.Region, s.PasswordID)
	if err != nil {
		return "", "", fmt.Errorf("retrieving password: %s", err)
	}

	return username, password, nil
}
