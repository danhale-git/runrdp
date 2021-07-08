package creds

import (
	"github.com/danhale-git/runrdp/internal/thycotic"
	"github.com/spf13/viper"
)

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
