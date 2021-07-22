package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/danhale-git/runrdp/internal/config/hosts"

	"github.com/danhale-git/runrdp/internal/config"

	"github.com/rgzr/sshtun"

	"github.com/danhale-git/runrdp/internal/rdp"

	"github.com/spf13/cobra"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var configuration *config.Configuration
var debug bool

// Execute begins execution of the CLI program
func Execute() {
	root := rootCommand()

	err := viper.BindPFlags(root.PersistentFlags())
	if err != nil {
		panic(err)
	}

	root.AddCommand(findCommand())
	root.AddCommand(versionCommand())

	vipers, err := readAllConfigs(viper.GetString("config-root"), ".toml")
	if err != nil {
		log.Fatal(err)
	}

	configuration, err = config.New(vipers)
	if err != nil {
		log.Fatalf("parsing configs: %s", err)
	}

	if err = root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func readAllConfigs(directory, extension string) (map[string]*viper.Viper, error) {
	infos, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	files := make([]*os.File, 0)
	for _, f := range infos {
		if filepath.Ext(f.Name()) == extension {
			f, err := os.Open(filepath.Join(directory, f.Name()))
			if err != nil {
				return nil, err
			}
			files = append(files, f)
		}
	}

	configs := make(map[string]io.Reader)
	for _, f := range files {
		configs[f.Name()] = f
	}

	vipers, err := config.ReadConfigs(configs)
	if err != nil {
		log.Fatalf("reading configs: %s", err)
	}

	for _, f := range files {
		if err := f.Close(); err != nil {
			fmt.Println("error closing file:", err)
		}
	}

	return vipers, nil
}

func rootCommand() *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	command := &cobra.Command{
		Use:   "runrdp",
		Short: "TBD",
		Long:  `TBD`,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.RangeArgs(1, 1)(cmd, args)
		},
		Run: Run,
	}

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	configRoot := filepath.Join(home, "/.runrdp/")

	// Desktop configuration
	command.PersistentFlags().BoolP("fullscreen", "f", false,
		"Starts Remote Desktop Connection in full-screen mode")
	command.PersistentFlags().BoolP("span", "s", false,
		"Matches the Remote Desktop width and height with the local virtual desktop, spanning across multiple monitors if necessary")
	/*command.PersistentFlags().BoolP("public", "p", false,
	"Runs Remote Desktop in public mode. In public mode, passwords and bitmaps aren't cached")*/

	command.PersistentFlags().Int("width", 0,
		"Specifies the width of the Remote Desktop window",
	)
	command.PersistentFlags().Int("height", 0,
		"Specifies the height of the Remote Desktop window",
	)

	command.PersistentFlags().String("address", "",
		"Hostname or IP address to connect to",
	)
	command.PersistentFlags().String("port", "",
		"Port to connect over",
	)

	command.PersistentFlags().StringP("username", "u", "",
		"Username to authenticate with",
	)
	command.PersistentFlags().StringP("password", "p", "",
		"Password to authenticate with",
	)

	// RunRDP config
	command.PersistentFlags().Bool("debug", false,
		"Print debug information",
	)

	command.PersistentFlags().String("tempfile-path", filepath.Join(configRoot, "connection.rdp"),
		"The directory in which a temporary .rdp file will be saved and run. Default is ~/.runrdp/",
	)

	command.PersistentFlags().String("config-root", configRoot,
		"directory containing config files",
	)

	command.PersistentFlags().String("ssh-directory", path.Join(home, ".ssh"),
		"Directory containing SSH keys.",
	)

	command.PersistentFlags().String("thycotic-url", path.Join(home, ""),
		"URL for Thycotic Secret Server.",
	)

	command.PersistentFlags().String("thycotic-domain", path.Join(home, ""),
		"Active Directory domain for Thycotic Secret Server.",
	)

	return command
}

// Run attempts to locate the given argument in the hosts config. If it is not a config entry the argument is validated
// as a socket and a connection is attempted if validation passes.
func Run(_ *cobra.Command, args []string) {
	// Config keys are always parsed to lower case.
	arg := strings.ToLower(args[0])

	if configuration.HostExists(arg) {
		connectToHost(arg)
		return
	} else {
		fmt.Printf("host %s does not exist in config\n", arg)
	}
}

func connectToHost(host string) {
	debug = viper.GetBool("debug")

	address, port := getSocket(host)

	username, password := getCredentials(host)

	if port == "" {
		port = rdp.DefaultPort
	}

	var tunnel *sshtun.SSHTun
	t, ok := configuration.Tunnels[configuration.HostGlobals[host][hosts.GlobalTunnel.String()]]
	if ok {
		tunnel, err := sshTunnel(&t, address, port)
		if err != nil {
			log.Fatalf("opening ssh tunnel: %s", err)
		}

		address = "localhost"
		port = t.LocalPort

		defer tunnel.Stop()
	}

	settings := getSettings(host)

	params := rdp.RDP{
		Username: username, Password: password,
		Address: address, Port: port,
		Width: settings.Width, Height: settings.Height,
		Fullscreen: settings.Fullscreen, Public: settings.Public, Span: settings.Span,
	}

	fmt.Printf("connecting to %s: %s:%s\n", host, address, port)

	if debug {
		b, err := json.MarshalIndent(params, "", "  ")
		if err != nil {
			log.Printf("error marshaling parameters for --debug: %s", err)
		} else {
			fmt.Println(strings.Replace(string(b), password, "REMOVED", 1))
		}
	}

	// Connect to the remote desktop.
	if err := rdp.Connect(&params, debug); err != nil {
		if tunnel != nil {
			tunnel.Stop()
		}
		log.Fatal(err)
	}

	// Close SSH connection when program exits. Wait for user to confirm before exiting.
	if tunnel != nil {
		fmt.Println("Press Enter to close SSH tunnel")
		if _, err := fmt.Scanln(); err != nil {
			log.Fatal(err)
		}
	}
}

func getSocket(host string) (string, string) {
	address, port, err := configuration.HostSocket(host, false)
	if err != nil {
		log.Fatalf("error getting host socket: %s", err)
	}

	if viper.GetString("address") != "" {
		address = viper.GetString("address")
	}

	if viper.GetString("port") != "" {
		port = viper.GetString("port")
	}

	return address, port
}

func getCredentials(host string) (string, string) {
	username, password, err := configuration.HostCredentials(host)
	if err != nil {
		fmt.Printf("error getting host credentials: %s\n", err)
	}

	if viper.GetString("username") != "" {
		username = viper.GetString("username")
	}

	if viper.GetString("password") != "" {
		password = viper.GetString("password")
	}

	return username, password
}

func getSettings(host string) config.Settings {
	name := configuration.HostGlobals[host][hosts.GlobalSettings.String()]
	if name == config.DefaultSettingsName {
		fmt.Printf("%s: 'settings = \"%s\"' can be removed because '%s' is the global default settings name.\n",
			host, config.DefaultSettingsName, config.DefaultSettingsName)
	}
	settings, ok := configuration.Settings[name]

	if !ok {
		dfault, ok := configuration.Settings[config.DefaultSettingsName]
		if !ok {
			settings = config.Settings{}
		} else {
			settings = dfault
		}
	}

	if viper.GetInt("height") != 0 {
		settings.Height = viper.GetInt("height")
	}
	if viper.GetInt("width") != 0 {
		settings.Width = viper.GetInt("width")
	}
	if !viper.GetBool("fullscreen") {
		settings.Fullscreen = viper.GetBool("fullscreen")
	}
	if !viper.GetBool("public") {
		settings.Public = viper.GetBool("public")
	}
	if !viper.GetBool("span") {
		settings.Span = viper.GetBool("span")
	}

	// Always disable public because there's no need to store configuration with runrdp
	settings.Public = false

	return settings
}

// sshTunnel open an SSH tunnel (port forwarding) equivalent to the command below:
//
// ssh -i <key file> -N -L <local port>:<host address>:<remote port> <username>@<forwarding server>
func sshTunnel(tunnel *config.Tunnel, address, port string) (*sshtun.SSHTun, error) {
	debug := viper.GetBool("debug")

	lp, err := strconv.Atoi(tunnel.LocalPort)
	if err != nil {
		return nil, fmt.Errorf("invalid local port '%s': %w", tunnel.LocalPort, err)
	}

	rp, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid remote port '%s': %w", port, err)
	}

	// Get the address of the intermediate host
	server, _, err := configuration.HostSocket(tunnel.Host, true)
	if err != nil {
		return nil, fmt.Errorf("getting ssh tunnel server address: %s", err)
	}

	sshTun := sshtun.New(lp, server, rp)
	sshTun.SetKeyFile(tunnel.Key)
	sshTun.SetUser(tunnel.User)
	sshTun.SetRemoteHost(address)

	// We enable debug messages to see what happens
	sshTun.SetDebug(debug) //DEBUG

	if debug {
		// Print the equivalent SSH command
		fmt.Printf("ssh -i %s -N -L %d:%s:%d %s@%s\n",
			tunnel.Key,
			lp,
			address,
			rp,
			tunnel.User,
			server,
		)

		// Print tunnel status changes
		sshTun.SetConnState(func(tun *sshtun.SSHTun, state sshtun.ConnState) {
			switch state {
			case sshtun.StateStarting:
				log.Printf("SSH tunnel starting")
			case sshtun.StateStarted:
				log.Printf("SSH tunnel open")
			case sshtun.StateStopped:
				log.Printf("SSH tunnel Stopped")
			}
		})
	}

	// Start the tunnel
	go func() {
		if err := sshTun.Start(); err != nil {
			log.Printf("SSH tunnel stopped: %s", err.Error())
		}
	}()

	return sshTun, nil
}
