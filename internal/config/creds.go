package config

import (
	"fmt"
	"os"

	"github.com/danhale-git/runrdp/internal/aws"
	"github.com/spf13/viper"
)

// EC2GetPassword is implements Cred and retrieves the EC2 initial administrator credentials for the given EC2Host.
type EC2GetPassword struct {
	Host *EC2Host // The EC2 host config
}

// Retrieve returns the administrator credentials for this instance or exists if unable to retrieve them.
func (p EC2GetPassword) Retrieve() (username, password string) {
	// Get instance password
	password, err := aws.GetPassword(
		p.Host.Profile,
		p.Host.Region,
		p.Host.ID,
		viper.GetString("ssh-directory"),
	)
	if err != nil {
		fmt.Println("Error retrieving EC2 Administrator password: ", err)
		os.Exit(1)
	}

	username = "Administrator"

	return
}
