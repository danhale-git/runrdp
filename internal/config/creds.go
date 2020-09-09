package config

import (
	"fmt"

	"github.com/danhale-git/runrdp/internal/aws"
	"github.com/spf13/viper"
)

// EC2GetPassword implements Cred and retrieves the EC2 initial administrator credentials for the given EC2Host.
type EC2GetPassword struct {
	Host *EC2Host // The EC2 host config
}

// Retrieve returns the administrator credentials for this instance or exists if unable to retrieve them.
func (p EC2GetPassword) Retrieve() (string, string, error) {
	// Get instance password
	password, err := aws.GetPassword(
		p.Host.Profile,
		p.Host.Region,
		p.Host.ID,
		viper.GetString("ssh-directory"),
	)
	if err != nil {
		return "", "", fmt.Errorf("getting ec2 administrator credentials: %s", err)
	}

	username := "Administrator"

	return username, password, nil
}

// SecretsManager implements Cred and retrieves a username and password from AWS Secrets Manager.
type SecretsManager struct {
	Username string
	Password string
	Region   string
}

// Retrieve returns the values for the configured Secrets Manager keys.
func (s SecretsManager) Retrieve() (string, string, error) {
	username, err := aws.GetSecret(s.Region, s.Username)
	if err != nil {
		return "", "", fmt.Errorf("retrieving username: %s", err)
	}

	password, err := aws.GetSecret(s.Region, s.Password)
	if err != nil {
		return "", "", fmt.Errorf("retrieving password: %s", err)
	}

	return username, password, nil
}
