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

// EC2PasswordCred implements Cred and retrieves the EC2 initial administrator credentials for the given EC2Host.
type EC2PasswordCred struct {
	Host *EC2Host // The EC2 host config
}

// Retrieve returns the administrator credentials for this instance or exists if unable to retrieve them.
func (p *EC2PasswordCred) Retrieve() (string, string, error) {
	svc := ec2instances.NewSession(p.Host.Profile, p.Host.Region)
	instance, err := ec2instances.InstanceFromID(svc, p.Host.ID)

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
		p.Host.ID,
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
	Username string
	Password string
	Profile  string
	Region   string
}

// Retrieve returns the values for the configured Secrets Manager keys.
func (s *SecretsManagerCred) Retrieve() (string, string, error) {
	username, err := secrets.Get(s.Profile, s.Region, s.Username)
	if err != nil {
		return "", "", fmt.Errorf("retrieving username: %s", err)
	}

	password, err := secrets.Get(s.Profile, s.Region, s.Password)
	if err != nil {
		return "", "", fmt.Errorf("retrieving password: %s", err)
	}

	return username, password, nil
}
