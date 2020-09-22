package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danhale-git/runrdp/internal/config"
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

			if !CheckExistence(
				path,
				fmt.Sprintf("%s config file", f),
				false,
			) {
				return
			}
			configuration.Data = map[string]*viper.Viper{f: configuration.Data[f]}
		}

		// No arguments given
		if len(args) == 0 {
			// List config entries
			if viper.GetBool("list") {
				fmt.Println("\nHosts:")
				for _, c := range configuration.Data {
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
			//name := viper.GetString("name")
			//fmt.Println(configuration.Get(name))
		default:
		}
	},
}

func checkDefaultConfig() bool {
	configRoot := viper.GetString("config-root")
	if !CheckExistence(
		configRoot,
		"config directory",
		true,
	) {
		return false
	}

	return CheckExistence(
		filepath.Join(configRoot, config.DefaultConfigName),
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

// CheckExistence checks for the existence of a file or directory, prompts the user to create it and returns true if
// it either already existed or the user chose to create it.
func CheckExistence(path, description string, dir bool) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("%s does not exist\ncreate at %s? (y/n): ", description, path)

		yes := interactiveYesNo()
		if yes {
			if dir {
				fmt.Println("created directory ", path)

				err = os.MkdirAll(path, 0600)
				if err != nil {
					fmt.Printf("failed to create directory %s: %s\n", path, err)
				}
			} else {
				fmt.Println("created file ", path)

				_, err = os.Create(path)
				if err != nil {
					fmt.Printf("failed to create file %s: %s\n", path, err)
				}
			}
			// User chose to create
			return true
		}
		// User chose not to create
		return false
	}
	// Item already exists
	return true
}

func interactiveYesNo() bool {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	return strings.TrimSpace(strings.ToLower(text)) == "y"
}
