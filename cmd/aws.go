package cmd

import (
	"fmt"
	"os"

	"github.com/danhale-git/runrdp/internal/aws"
	"github.com/danhale-git/runrdp/internal/rdp"

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
	// Create AWS API session
	sess := aws.NewSession(
		viper.GetString("profile"),
		viper.GetString("region"),
	)

	// Get instance
	instance, err := aws.InstanceFromID(sess, viper.GetString("ec2-id"))
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// Get instance password
	password, err := aws.GetPassword(sess, viper.GetString("ec2-id"), "")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	user := "Administrator"
	host := instance.PrivateIpAddress

	rdp.Connect(*host, user, password)
}

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().StringP("profile", "p", "", "AWS Config profile name.")
	awsCmd.MarkFlagRequired("profile")

	awsCmd.Flags().StringP("region", "r", "", "AWS region.")
	awsCmd.MarkFlagRequired("region")

	awsCmd.Flags().StringP("ec2-id", "i", "", "AWS EC2 instance ID.")
	awsCmd.MarkFlagRequired("ec2-id")

	err := viper.BindPFlags(awsCmd.Flags())
	if err != nil {
		panic(err)
	}
}
