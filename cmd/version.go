package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.1.1"

func versionCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "version",
		Short: "Print the current version of runrdp",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(version)
		},
	}

	return command
}
