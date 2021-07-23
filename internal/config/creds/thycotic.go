package creds

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/danhale-git/runrdp/internal/thycotic"
	"github.com/danhale-git/tss-sdk-go/server"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

// ThycoticStruct a struct of type creds.Thycotic.
func ThycoticStruct() interface{} {
	return &Thycotic{}
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
		return "", "", fmt.Errorf("invalid thycotic url '%s': %w", err)
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

	return thycotic.GetCredentials(s, t.SecretID)
}

func credentials() (string, string, error) {
	stdin := os.Stdin
	reader := bufio.NewReader(stdin)

	//defer stdin.Close()

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
