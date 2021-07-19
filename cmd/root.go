package cmd

import (
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

// Execute begins execution of the CLI program
func Execute() {
	root := rootCommand()

	err := viper.BindPFlags(root.PersistentFlags())
	if err != nil {
		panic(err)
	}

	root.AddCommand(findCommand())

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

	command.PersistentFlags().Bool("debug", false,
		"Print debug information",
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

	// Check if command line flags were passed and override configuration
	clAddress := viper.GetString("address")
	if clAddress != "" {
		address = clAddress
	}

	clPort := viper.GetString("port")
	if clPort != "" {
		port = clPort
	}

	// If no port was given, use the standard RDP port.
	if port == "" {
		port = rdp.DefaultPort
	}

	socket := fmt.Sprintf("%s:%s", address, port)

	// Open an ssh tunnel and replace socket with the localhost:localport tunnel socket.
	var tunnel *sshtun.SSHTun
	t, ok := configuration.Tunnels[configuration.HostGlobals[host][hosts.GlobalTunnel.String()]]
	if ok {
		tunnel, err = sshTunnel(&t, address, port)
		if err != nil {
			log.Fatalf("opening ssh tunnel: %s", err)
		}

		socket = fmt.Sprintf("localhost:%s", t.LocalPort)

		defer tunnel.Stop()
	}

	settings, ok := configuration.Settings[configuration.HostGlobals[host][hosts.GlobalSettings.String()]]
	if !ok {
		settings = config.Settings{}
	}

	// Connect to the remote desktop.
	rdp.Connect(
		socket,
		username,
		password,
		viper.GetString("tempfile-path"),
		settings.Width,
		settings.Height,
		settings.Scale,
	)

	// Close SSH connection when program exits. Wait for user to confirm before exiting.
	if tunnel != nil {
		fmt.Println("Press Enter to close SSH tunnel")
		if _, err := fmt.Scanln(); err != nil {
			log.Fatal(err)
		}
	}
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
