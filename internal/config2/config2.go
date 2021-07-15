package config2

import (
	"fmt"
	"io"

	"github.com/danhale-git/runrdp/internal/config2/creds"

	"github.com/danhale-git/runrdp/internal/config2/hosts"

	"github.com/spf13/viper"
)

type Configuration struct {
	//Data        map[string]*viper.Viper      // Data from individual config files
	Hosts       map[string]hosts.Host        // Host configs
	HostGlobals map[string]map[string]string // Global Host fields by [host key][field name]

	creds    map[string]creds.Cred `mapstructure:"cred"`
	tunnels  map[string]Tunnel     `mapstructure:"tunnel"`
	settings map[string]Settings   `mapstructure:"setting"`
}

// Tunnel has the details for opening an 'SSH tunnel' (SSH port forwarding) including a reference to a Host config which
// will be the forwarding server.
type Tunnel struct {
	Host      string `mapstructure:"host"`
	LocalPort string `mapstructure:"localport"`
	Key       string `mapstructure:"key"`
	User      string `mapstructure:"user"`
}

func (t Tunnel) Validate() error {
	return nil
}

// Settings is the configuration of .RDP file settings.
// https://docs.microsoft.com/en-us/windows-server/remote/remote-desktop-services/clients/rdp-files
type Settings struct {
	Height int `mapstructure:"height"`
	Width  int `mapstructure:"width"`
	Scale  int `mapstructure:"scale"`
}

func (s Settings) Validate() error {
	if s.Width != 0 && (s.Width < 200 || s.Width > 8192) {
		return fmt.Errorf("width value is %d invalid, must be above 200 and below 8192\n", s.Width)
	}

	if s.Height != 0 && (s.Height < 200 || s.Height > 8192) {
		return fmt.Errorf("height value %d is invalid, must be above 200 and below 8192\n", s.Height)
	}

	if s.Scale != 0 && func() bool {
		// Scale is not in list of valid values
		for _, v := range []int{100, 125, 150, 175, 200, 250, 300, 400, 500} {
			if s.Scale == v {
				return false
			}
		}
		return true
	}() {
		return fmt.Errorf("scale value %d is invalid, must be one of 100, 125, 150, 175, 200, 250, 300, 400, 500\n", s.Scale)
	}

	return nil
}

// ReadConfigs reads a map of io.Reader into a matching map of viper.Viper.
func ReadConfigs(readers map[string]io.Reader) (map[string]*viper.Viper, error) {
	vipers := make(map[string]*viper.Viper)
	for k, r := range readers {
		v := viper.New()
		v.SetConfigType("toml")
		if err := v.ReadConfig(r); err != nil {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		vipers[k] = v
	}

	return vipers, nil
}

// New takes a map of viper instances and parses them to a Configuration struct.
func New(v map[string]*viper.Viper) (*Configuration, error) {
	c := Configuration{}

	c.Hosts = make(map[string]hosts.Host)
	c.HostGlobals = make(map[string]map[string]string)
	c.creds = make(map[string]creds.Cred)
	c.tunnels = make(map[string]Tunnel)
	c.settings = make(map[string]Settings)

	if err := parseHosts(v, c.Hosts, c.HostGlobals); err != nil {
		return nil, fmt.Errorf("parsing hosts: %w", err)
	}

	if err := parseCreds(v, c.creds); err != nil {
		return nil, fmt.Errorf("parsing creds: %w", err)
	}

	if err := parseSettings(v, c.settings); err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}

	if err := parseTunnels(v, c.tunnels); err != nil {
		return nil, fmt.Errorf("parsing tunnels: %w", err)
	}

	return &c, nil
}
