package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/danhale-git/runrdp/internal/desktops"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rdp",
	Short: "TBD",
	Long:  `TBD`,
	Run:   func(cmd *cobra.Command, args []string) {},
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

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.rdp.yaml)")
	rootCmd.PersistentFlags().StringP("desktops", "d", "", "desktop file (default is $HOME/.rdp.desktops.yaml)")

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
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

	fmt.Println(viper.GetString("Test"))

	config := loadDesktopConfig(home)

	desktops.LoadDesktops(config)
}

func loadDesktopConfig(home string) desktops.Config {
	var c desktops.Config

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

	return c
}

/*func loadDesktopConfig(home string) desktops.Config {
	var c desktops.Config
	desktopFile := viper.GetString("desktops")

	if desktopFile == "" {
		desktopFile = path.Join(home, ".desktops.yaml")
	}

	yamlFile, err := ioutil.ReadFile(desktopFile)
	if err != nil {
		log.Printf("error reading yaml file: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("invalid desktops config: %v", err)
	}

	fmt.Println(string(yamlFile))
	fmt.Println("Found desktops:", desktopFile)

	return c
}*/
