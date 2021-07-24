package thycotic

import (
	"fmt"

	"github.com/danhale-git/tss-sdk-go/server"
)

// Secreter returns a secret.
type Secreter interface {
	Secret(id int) (*server.Secret, error)
}

// NewServer builds a configuration and returns a new Thycotic server.
func NewServer(user, pass, u, d string) (*server.Server, error) {
	c := server.Configuration{
		Credentials: server.UserCredential{
			Username: user,
			Password: pass,
		},
		Domain:    d,
		ServerURL: u,
	}

	return server.New(c)
}

// GetCredentials calls the Thycotic API via the Go SDK, obtains a secret and attempts to get the Username and Password
// fields. If either field is not present an error is returned.
func GetCredentials(s Secreter, secretID int) (string, string, error) {
	secret, err := s.Secret(secretID)

	if err != nil {
		return "", "", fmt.Errorf("getting secret '%d' from thycotic: %s", secretID, err)
	}

	username, err := getField(secret, "Username")
	if err != nil {
		return "", "", err
	}

	password, err := getField(secret, "Password")
	if err != nil {
		return "", "", err
	}

	return username, password, nil
}

func getField(secret *server.Secret, fieldName string) (string, error) {
	value, exists := secret.Field(fieldName)

	if !exists {
		return "", fmt.Errorf(
			"secret '%s' with template ID %d has no field '%s'",
			secret.Name,
			secret.SecretTemplateID,
			fieldName,
		)
	}

	return value, nil
}
