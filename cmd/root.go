package cmd

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/danhale-git/runrdp/internal/rdp"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var desktops rdp.Desktops

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rdp",
	Short: "TBD",
	Long:  `TBD`,
	Args: func(cmd *cobra.Command, args []string) error {
		cobra.RangeArgs(1, 1)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		arg := args[0]
		d, ok := desktops[arg]
		if ok {
			username, password := d.Credentials.Retrieve()
			rdp.Connect(d.Host.Socket(), username, password)
		} else if host, err := net.LookupHost(arg); err == nil {
			rdp.Connect(host[0], "", "")
		}

		fmt.Printf("'%s' is not a config entry, hostname or IP address", arg)
		os.Exit(0)
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

	rootCmd.PersistentFlags().StringP("username", "u", "", "Username to authenticate with")
	rootCmd.PersistentFlags().StringP("password", "p", "", "Password to authenticate with")

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.rdp.yaml)")
	rootCmd.PersistentFlags().StringP("desktops", "d", "", "desktop file (default is $HOME/.rdp.desktops.yaml)")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		panic(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configFile := viper.GetString("config")

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(home)
		viper.SetConfigName(".rdp")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Found config:", viper.ConfigFileUsed())
	}

	fmt.Println("Test config string: ", viper.GetString("Test"))

	config := loadDesktopConfig(home)

	desktops = rdp.LoadDesktops(config)
}

func loadDesktopConfig(home string) []rdp.DesktopConfig {
	var c struct{ Desktops []rdp.DesktopConfig }

	desktopViper := viper.New()
	desktopFile := viper.GetString("desktops")

	if desktopFile != "" {
		desktopViper.SetConfigFile(desktopFile)
	} else {
		desktopViper.AddConfigPath(home)
		desktopViper.SetConfigName(".desktops")
	}

	// If a config file is found, read it in.
	if err := desktopViper.ReadInConfig(); err == nil {
		fmt.Println("Found desktops:", viper.ConfigFileUsed())
	}

	err := desktopViper.Unmarshal(&c)
	if err != nil {
		log.Printf("invalid desktops config: %v", err)
	}

	return c.Desktops
}
