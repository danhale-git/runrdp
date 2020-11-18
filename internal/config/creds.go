package config

import (
	"fmt"
	"reflect"

	"github.com/danhale-git/runrdp/internal/thycotic"

	"github.com/danhale-git/runrdp/internal/aws/ec2instances"
	"github.com/danhale-git/runrdp/internal/aws/secrets"

	"github.com/spf13/viper"
)

// GetCredential returns a credential struct and it's reflect.Value based on the given type key.
func GetCredential(key string) (Cred, reflect.Value, error) {
	switch key {
	case "awssm":
		cred := SecretsManagerCred{}
		return &cred, reflect.ValueOf(&cred).Elem(), nil
	case "thycotic":
		cred := ThycoticCred{}
		return &cred, reflect.ValueOf(&cred).Elem(), nil
	default:
		return nil, reflect.Value{},
			fmt.Errorf("cred type key'%s' not recognized", key)
	}
}

// Retrieve returns the administrator credentials for this instance or an error if unable to retrieve them. If the
// getcred field is not set it returns empty strings and no error.
func (h *EC2Host) Retrieve() (string, string, error) {
	if !h.GetCred {
		return "", "", nil
	}

	instance, err := h.Instance()

	if err != nil {
		return "", "", fmt.Errorf("getting ec2 instance: %s", err)
	}

	keyData, err := ec2instances.ReadPrivateKey(
		viper.GetString("ssh-directory"),
		*instance.KeyName,
	)

	if err != nil {
		return "", "", fmt.Errorf("reading private key: %s", err)
	}

	// Get instance password
	password, err := ec2instances.GetPassword(
		h.svc,
		*instance.InstanceId,
		keyData,
	)
	if err != nil {
		return "", "", fmt.Errorf("getting ec2 administrator credentials: %s", err)
	}

	username := "Administrator"

	return username, password, nil
}

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

// ThycoticCred implements cred and retrieves a username and password from Thycotic Secret Server.
type ThycoticCred struct {
	SecretID int
}

// Retrieve returns the username and password fields from the secret with the given ID. If either the 'Username' or
// 'Password' field not in the secret template an error is returned.
func (t *ThycoticCred) Retrieve() (string, string, error) {
	return thycotic.GetCredentials(
		t.SecretID,
		viper.GetString("thycotic-url"),
		viper.GetString("thycotic-domain"),
	)
}
