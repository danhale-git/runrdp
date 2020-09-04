package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/danhale-git/runrdp/internal/rdp"

	"github.com/danhale-git/runrdp/internal/aws"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// Session from shared creds: https://github.com/aws/aws-sdk-go/blob/v1.34.9/aws/session/session.go#L277
//https://docs.aws.amazon.com/sdk-for-go/api/aws/session/#Options

// Run gimme creds ad hoc https://golang.org/pkg/os/exec/#Cmd.Wait

// Get key name from instance: ec2.Instance.KeyName

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "TBD",
	Long:  `TBD`,
	Run:   awsCmdRun,
}

func awsCmdRun(cmd *cobra.Command, args []string) {
	host := aws.EC2Host{
		ID:      viper.GetString("ec2-id"),
		Private: viper.GetBool("private"),
		Profile: viper.GetString("profile"),
		Region:  viper.GetString("region"),
	}

	var credentials rdp.Credentials
	if viper.GetBool("awspass") {
		credentials = aws.EC2GetPassword{EC2Host: &host}
	}

	username, password := credentials.Retrieve()

	rdp.Connect(host.Socket(), username, password)
}

func init() {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().StringP("profile", "p", "default", "AWS Config profile name.")

	awsCmd.Flags().StringP("region", "r", "", "AWS region.")

	awsCmd.Flags().StringP("ec2-id", "i", "", "AWS EC2 instance ID.")
	awsCmd.MarkFlagRequired("ec2-id")

	awsCmd.Flags().Bool("private", false, "Use private IP address.")
	awsCmd.Flags().Bool("awspass", false, "Use private IP address.")

	viper.SetDefault("SSHDirectory", path.Join(home, ".ssh"))

	err = viper.BindPFlags(awsCmd.Flags())
	if err != nil {
		panic(err)
	}
}
