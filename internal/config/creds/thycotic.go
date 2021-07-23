package creds

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/danhale-git/tss-sdk-go/server"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

// ThycoticStruct returns a struct of type creds.Thycotic.
func ThycoticStruct() interface{} {
	return &Thycotic{}
}

type Secreter interface {
	Secret(id int) (*server.Secret, error)
}

// Validate returns an error if a config field is invalid.
func (t Thycotic) Validate() error {
	return nil
}

// Thycotic implements cred and retrieves a username and password from Thycotic Secret Server.
type Thycotic struct {
	SecretID int
}

// Retrieve returns the username and password fields from the secret with the given ID. If either the 'Username' or
// 'Password' field not in the secret template an error is returned.
func (t *Thycotic) Retrieve() (string, string, error) {
	thyUser, thyPassword, err := credentials()
	if err != nil {
		return "", "", err
	}

	d := viper.GetString("thycotic-domain")
	u := viper.GetString("thycotic-url")
	if _, err := url.Parse(u); err != nil {
		return "", "", fmt.Errorf("invalid thycotic url '%s': %w", u, err)
	}

	fmt.Println("Thycotic URL:", u, "Thycotic domain:", d)

	c := server.Configuration{
		Credentials: server.UserCredential{
			Username: thyUser,
			Password: thyPassword,
		},
		Domain:    d,
		ServerURL: u,
	}

	s, err := server.New(c)
	if err != nil {
		return "", "", fmt.Errorf("creating thycotic server: %s", err)
	}

	return GetCredentials(s, t.SecretID)
}

func credentials() (string, string, error) {
	stdin := os.Stdin
	reader := bufio.NewReader(stdin)

	defer stdin.Close()

	fmt.Print("Enter Thycotic Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Thycotic Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("reading password input: %s", err)
	}

	fmt.Println()

	password := string(bytePassword)

	return strings.TrimSpace(username), strings.TrimSpace(password), nil
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
