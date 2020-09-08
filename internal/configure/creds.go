package configure

import (
	"fmt"
	"os"

	"github.com/danhale-git/runrdp/internal/aws"
	"github.com/spf13/viper"
)

type EC2GetPassword struct {
	Host *EC2Host
}

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
