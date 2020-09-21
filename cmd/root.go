package cmd

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/danhale-git/runrdp/internal/config"

	"github.com/danhale-git/runrdp/internal/rdp"

	"github.com/spf13/cobra"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var configuration config.Configuration

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rdp",
	Short: "TBD",
	Long:  `TBD`,
	Args: func(cmd *cobra.Command, args []string) error {
		return cobra.RangeArgs(1, 1)(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		arg := args[0]
		host, ok := configuration.Hosts[arg]
		if ok {
			var address string
			var err error
			// If a proxy was defined, use it's address
			if configuration.Proxys[arg] != nil {
				p := *configuration.Proxys[arg]
				address, err = p.GetAddress()
			} else {
				address, err = host.GetAddress()
			}

			if err != nil {
				fmt.Printf("Error retrieving host address: %s\n", err)
				return
			}

			var username, password string

			if cred := host.Credentials(); cred != nil {
				username, password, err = cred.Retrieve()
			} else if configuration.Creds[arg] != nil {
				c := *configuration.Creds[arg]
				username, password, err = c.Retrieve()
			} else {
				username, password = "", ""
			}

			if err != nil {
				fmt.Printf("Error retrieving credentials: %s\n", err)
			}

			p := host.GetPort()
			if p > 0 {
				address = fmt.Sprintf("%s:%d", address, p)
			}

			rdp.Connect(address, username, password)

		} else {
			split := strings.Split(arg, ":")
			_, err := net.LookupHost(split[0])
			if err == nil {
				rdp.Connect(arg, "", "")
			} else {
				fmt.Printf("'%s' is not a config entry, hostname or IP address\n", arg)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	configRoot := filepath.Join(home, "/.runrdp/")

	rootCmd.PersistentFlags().StringP(
		"username", "u", "",
		"Username to authenticate with",
	)
	rootCmd.PersistentFlags().StringP(
		"password", "p", "",
		"Password to authenticate with",
	)

	rootCmd.PersistentFlags().String(
		"config-root", configRoot,
		"directory containing config files",
	)

	rootCmd.PersistentFlags().String(
		"tag-separator", ";",
		"separator character for tags",
	)

	err = viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		panic(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Read the file 'config' into the default viper if it exists
	loadMainConfig()

	// Read all config files into separate viper instances
	// This includes 'config' which is read a second time here so it may include host and cred configurations
	configuration.ReadConfigFiles()
	configuration.BuildData()
}

func loadMainConfig() {
	root := viper.GetString("config-root")
	filePath := viper.GetString(DefaultConfigName)

	if filePath != "" {
		viper.SetConfigFile(filePath)
	} else {
		viper.AddConfigPath(root)
		viper.SetConfigName(DefaultConfigName)
	}

	viper.SetConfigType("toml")
	viper.SetConfigFile(filepath.Join(
		viper.GetString("config-root"),
		"config",
	))

	// No default config is required so do nothing if it isn't found
	_ = viper.ReadInConfig()
}
