package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/danhale-git/runrdp/internal/configure"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:       "configure",
	Short:     "TBD",
	Long:      `TBD`,
	ValidArgs: []string{"add", "show"},
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.OnlyValidArgs(cmd, args); err != nil {
			return err
		}

		if len(args) == 0 {
			return nil
		}

		switch args[0] {
		case "add":
		//stuff
		case "show":
			cmd.MarkFlagRequired("name")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Check config directory exists
		if !checkDefaultConfig() {
			return
		}

		// Limit operations to specific file name if given
		if viper.IsSet("file") {
			f := viper.GetString("file")
			path := filepath.Join(viper.GetString("config-root"), f)

			if !configure.CheckExistence(
				path,
				fmt.Sprintf("%s config file", f),
				false,
			) {
				return
			}
			config.Files = map[string]*viper.Viper{f: config.Files[f]}
		}

		// No arguments given
		if len(args) == 0 {
			// List config entries
			if viper.GetBool("list") {
				fmt.Println("\nHosts:")
				for _, c := range config.Files {
					for kind, names := range c.GetStringMap("host") {
						for name := range names.(map[string]interface{}) {
							fmt.Printf("  %s (%s)", name, kind)
						}
						fmt.Println()
					}
				}
				return
			}

			return
		}

		switch args[0] {
		case "add":

			// configure something
			/*configuration, name, ok := configure.AddCredentialsInteractive()
			if !ok {
				return
			}

			// choose config file
			config.Set(name, configuration)

			err := config.WriteConfigAs(path)
			if err != nil {
				fmt.Println(err)
			}*/

		case "show":
			name := viper.GetString("name")
			fmt.Println(config.Get(name))
		default:
		}
	},
}

func checkDefaultConfig() bool {
	configRoot := viper.GetString("config-root")
	if !configure.CheckExistence(
		configRoot,
		"config directory",
		true,
	) {
		return false
	}

	return configure.CheckExistence(
		filepath.Join(configRoot, configure.DefaultConfigName),
		"default config file",
		false,
	)
}

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.Flags().BoolP("list", "l", false, "list config entries")
	configureCmd.Flags().StringP("name", "n", "", "full name of the config entry to operate on")
	configureCmd.Flags().StringP("file", "f", "", "name of the config file to operate on")

	viper.BindPFlags(configureCmd.Flags())
}
