package cmd

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
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
	Run: Run,
}

// Run attempts to locate the given argument in the hosts config. If it is not a config entry the argument is validated
// as a socket and a connection is attempted if validation passes.
func Run(_ *cobra.Command, args []string) {
	// Config keys are always parsed to lower case.
	arg := strings.ToLower(args[0])

	if configuration.HostExists(arg) {
		connectToHost(arg)
		return
	}

	if SocketArgument(arg) {
		return
	}

	fmt.Println(arg, "is not a hostname, ip address or config key.")
}

func connectToHost(host string) {
	fmt.Printf("Connecting to: %s\n", host)

	// Get the host socket from configuration.
	address, port, err := configuration.HostSocket(host, false)

	if err != nil {
		log.Fatalf("error getting host socket: %s", err)
	}

	var username, password string
	username, password, err = configuration.HostCredentials(host)

	if err != nil {
		fmt.Printf("error getting host credentials: %s\n", err)
	}

	clAddress := viper.GetString("address")

	if clAddress != "" {
		// Address command line flag was passed, override configuration.
		address = clAddress
	}

	clPort := viper.GetString("port")

	if clPort != "" {
		// Port command line flag was passed, override configuration.
		port = clPort
	}

	if port == "" {
		// No port was given, use the standard RDP port.
		port = rdp.DefaultPort
	}

	socket := fmt.Sprintf("%s:%s", address, port)

	t, err := configuration.HostTunnel(host)
	if err != nil {
		log.Fatal(err)
	}

	// Open an ssh tunnel and replace socket with the localhost:localport tunnel socket.
	if t != nil {
		socket, err = sshTunnel(t, address, port)
		if err != nil {
			log.Fatalf("opening ssh tunnel: %s", err)
		}
	}

	rdp.Connect(socket, username, password)
}

// sshTunnel runs the following ssh command using exec:
//
// ssh -i <key file> -N -L <local port>:<host address>:<remote port> <username>@<forwarding server>
func sshTunnel(tunnel *config.SSHTunnel, address, port string) (string, error) {
	server, _, err := configuration.HostSocket(tunnel.Host, true)
	if err != nil {
		return "", fmt.Errorf("getting ssh tunnel server address: %s", err)
	}

	t := fmt.Sprintf("%s:%s:%s", tunnel.LocalPort, address, port)
	u := fmt.Sprintf("%s@%s", tunnel.User, server)

	command := exec.Command("ssh", "-i", tunnel.Key, "-N", "-L", t, u)

	fmt.Printf("ssh tunnel open %s %s\n", t, u)

	err = command.Start()
	if err != nil {
		return "", fmt.Errorf("starting command '%s': %s", command.String(), err)
	}

	go func() {
		cmdErr := command.Wait()
		if cmdErr != nil {
			fmt.Printf("ssh tunnel command exited with error: %s\n", cmdErr)
		}

		fmt.Println("ssh tunnel closed")
	}()

	return fmt.Sprintf("localhost:%s", tunnel.LocalPort), nil
}

// SocketArgument checks if arg is 'host' or 'host:port' and attempts to connect if it is. It returns a bool indicating
// whether it attempted to connect
func SocketArgument(arg string) bool {
	var address, port string

	if strings.Contains(arg, ":") {
		split := strings.Split(arg, ":")
		address = split[0]

		if len(split) > 1 {
			port = split[1]
		}
	} else {
		address = arg
	}

	if ValidateAddress(address) {
		var socket string
		if port != "" {
			socket = fmt.Sprintf("%s:%s", address, port)
		} else {
			socket = address
		}

		rdp.Connect(
			socket,
			viper.GetString("username"),
			viper.GetString("password"),
		)

		return true
	}

	return false
}

// ValidateAddress returns true if the given string parses to an IP address or resolves an IP address in the DNS.
func ValidateAddress(address string) bool {
	ip := net.ParseIP(address)
	if ip != nil {
		return true
	} else if h, _ := net.LookupHost(address); len(h) > 0 {
		return true
	}

	return false
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

	rootCmd.PersistentFlags().String(
		"address", "",
		"Hostname or IP address to connect to",
	)
	rootCmd.PersistentFlags().String(
		"port", "",
		"Port to connect over",
	)

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

	rootCmd.PersistentFlags().String(
		"ssh-directory",
		path.Join(home, ".ssh"),
		"Directory containing SSH keys.",
	)

	rootCmd.PersistentFlags().String(
		"thycotic-url",
		path.Join(home, ""),
		"URL for Thycotic Secret Server.",
	)

	rootCmd.PersistentFlags().String(
		"thycotic-domain",
		path.Join(home, ""),
		"Active Directory domain for Thycotic Secret Server.",
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

	err := validateMainConfig()

	if err != nil {
		log.Fatalf("Default config is invalid: %s", err)
	}

	// Read all config files into separate viper instances
	// This includes 'config' which is read a second time here so it may include host and cred configurations
	configuration.ReadConfigFiles()
	configuration.BuildData()
}

func loadMainConfig() {
	root := viper.GetString("config-root")
	filePath := viper.GetString(config.DefaultConfigName)

	if filePath != "" {
		viper.SetConfigFile(filePath)
	} else {
		viper.AddConfigPath(root)
		viper.SetConfigName(config.DefaultConfigName)
	}

	viper.SetConfigType("toml")
	viper.SetConfigFile(filepath.Join(
		viper.GetString("config-root"),
		config.DefaultConfigName,
	))

	// No default config is required so do nothing if it isn't found
	_ = viper.ReadInConfig()
}

func validateMainConfig() error {
	for _, key := range viper.AllKeys() {
		topLevelKey := strings.Split(key, ".")[0]
		if topLevelKey == "host" || topLevelKey == "cred" {
			return fmt.Errorf("'%s': default config file ('%s') may not contain host or cred entries: run "+
				"`runrdp configure`", key, config.DefaultConfigName)
		}
	}

	return nil
}
