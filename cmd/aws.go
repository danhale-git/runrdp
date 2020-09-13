package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/danhale-git/runrdp/internal/config"

	"github.com/danhale-git/runrdp/internal/rdp"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "TBD",
	Long:  `TBD`,
	Run:   awsCmdRun,
}

func awsCmdRun(_ *cobra.Command, _ []string) {
	host := config.EC2Host{
		ID:      viper.GetString("ec2-id"),
		Private: viper.GetBool("private"),
		Profile: viper.GetString("profile"),
		Region:  viper.GetString("region"),
	}

	var credentials config.Cred

	var username, password string

	var err error

	if viper.GetBool("getpass") {
		credentials = &config.EC2PasswordCred{Host: &host}
		username, password, err = credentials.Retrieve()

		if err != nil {
			fmt.Printf("Error retrieving credentials: %s\n", err)
		}
	}

	socket, err := host.Socket()
	if err == nil {
		rdp.Connect(socket, username, password)
	} else {
		fmt.Printf("Error retrieving host address: %s\n", err)
	}
}

func init() {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().String("profile", "default", "AWS Config profile name.")

	awsCmd.Flags().StringP("region", "r", "", "AWS region.")

	awsCmd.Flags().StringP("ec2-id", "i", "", "AWS EC2 instance ID.")
	_ = awsCmd.MarkFlagRequired("ec2-id")

	awsCmd.Flags().Bool("private", false, "Use private IP address.")
	awsCmd.Flags().Bool("getpass", false, "Use private IP address.")

	awsCmd.Flags().String("ssh-directory", path.Join(home, ".ssh"), "Directory containing SSH keys.")

	err = viper.BindPFlags(awsCmd.Flags())
	if err != nil {
		panic(err)
	}
}
