package config

import (
	"fmt"
	"reflect"

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
	default:
		return nil, reflect.Value{},
			fmt.Errorf("cred type key'%s' not recognized", key)
	}
}

// Retrieve returns the administrator credentials for this instance or exists if unable to retrieve them.
func (h *EC2Host) Retrieve() (string, string, error) {
	if !h.GetCred {
		return "", "", nil
	}

	svc := ec2instances.NewSession(h.Profile, h.Region)
	instance, err := ec2instances.GetInstance(
		svc,
		h.ID,
		viper.GetString("tag-separator"),
		h.IncludeTags,
		h.ExcludeTags,
	)

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
		svc,
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
