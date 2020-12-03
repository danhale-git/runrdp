package thycotic

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/danhale-git/tss-sdk-go/server"
	"golang.org/x/crypto/ssh/terminal"
)

// GetCredentials calls the Thycotic API via the Go SDK, obtains a secret and attempts to get the Username and Password
// fields. If either field is not present an error is returned.
func GetCredentials(secretID int, url, domain string) (string, string, error) {
	thyUser, thyPassword, err := credentials()
	if err != nil {
		return "", "", err
	}

	c := server.Configuration{
		Credentials: server.UserCredential{
			thyUser,
			thyPassword,
		},
		ServerURL: url,
		Domain:    domain,
	}

	s, err := server.New(c)

	if err != nil {
		return "", "", fmt.Errorf("creating thycotic server: %s", err)
	}

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

func credentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

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
