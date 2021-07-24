package creds

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/danhale-git/runrdp/internal/config/creds/thycotic"

	"github.com/spf13/viper"
	"golang.org/x/term"
)

// ThycoticStruct returns a struct of type creds.Thycotic.
func ThycoticStruct() interface{} {
	return &Thycotic{}
}

// Thycotic implements cred and retrieves a username and password from Thycotic Secret Server.
type Thycotic struct {
	SecretID int
}

// Validate returns an error if a config field is invalid.
func (t Thycotic) Validate() error {
	if t.SecretID == 0 {
		return fmt.Errorf("invalid thycotic secret id: 0")
	}

	return nil
}

// Retrieve returns the username and password fields from the secret with the given ID. If either the 'Username' or
// 'Password' field not in the secret template an error is returned.
func (t *Thycotic) Retrieve() (string, string, error) {
	user, pass, err := credentials()
	if err != nil {
		return "", "", err
	}

	d := viper.GetString("thycotic-domain")
	u := viper.GetString("thycotic-url")
	if _, err := url.Parse(u); err != nil {
		return "", "", fmt.Errorf("invalid thycotic url '%s': %w", u, err)
	}

	s, err := thycotic.NewServer(user, pass, u, d)
	if err != nil {
		return "", "", fmt.Errorf("creating thycotic server: %s", err)
	}

	return thycotic.GetCredentials(s, t.SecretID)
}

func credentials() (string, string, error) {
	stdin := os.Stdin
	reader := bufio.NewReader(stdin)

	fmt.Print("Enter Thycotic Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Thycotic Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("reading password input: %s", err)
	}

	fmt.Println()

	password := string(bytePassword)

	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}
