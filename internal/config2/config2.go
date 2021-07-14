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

// Settings is the configuration of .RDP file settings.
// https://docs.microsoft.com/en-us/windows-server/remote/remote-desktop-services/clients/rdp-files
type Settings struct {
	Height int `mapstructure:"height"`
	Width  int `mapstructure:"width"`
	Scale  int `mapstructure:"scale"`
}

/*if settings.Width != 0 && (settings.Width < 200 || settings.Width > 8192) {
fmt.Printf("Failed to load settings '%s = %d': width value is invalid, must be above 200 and below 8192\n", itemKey, settings.Width)
continue
}

if settings.Height != 0 && (settings.Height < 200 || settings.Height > 8192) {
fmt.Printf("Failed to load settings '%s = %d': height value is invalid, must be above 200 and below 8192\n", itemKey, settings.Height)
continue
}

if settings.Scale != 0 && func() bool {
	// Scale is not in list of valid values
	for _, v := range []int{100, 125, 150, 175, 200, 250, 300, 400, 500} {
		if settings.Scale == v {
			return false
		}
	}
	return true
}() {
fmt.Printf("Failed to load settings '%s': scale value is invalid, must be one of 100, 125, 150, 175, 200, 250, 300, 400, 500\n", itemKey)
continue
}*/

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

	for key, typeFunc := range hosts.Map {
		h, err := parse(v, fmt.Sprintf("host.%s", key), typeFunc)
		if err != nil {
			return nil, fmt.Errorf("parsing hosts: %w", err)
		}
		for k, v := range h {
			if _, ok := c.Hosts[k]; ok {
				return nil, &DuplicateConfigNameError{Name: k}
			}
			c.Hosts[k] = v.(hosts.Host)
		}
		g, err := parseGlobals(v, fmt.Sprintf("host.%s", key))
		for k, v := range g {
			c.HostGlobals[k] = v.(map[string]string)
		}
	}

	for key, typeFunc := range creds.Map {
		cr, err := parse(v, fmt.Sprintf("cred.%s", key), typeFunc)
		if err != nil {
			return nil, fmt.Errorf("parsing creds: %w", err)
		}
		for k, v := range cr {
			if _, ok := c.creds[k]; ok {
				return nil, &DuplicateConfigNameError{Name: k}
			}
			c.creds[k] = v.(creds.Cred)
		}
	}

	s, err := parse(v, "settings", func() interface{} { return &Settings{} })
	if err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}
	for k, v := range s {
		if _, ok := c.settings[k]; ok {
			return nil, &DuplicateConfigNameError{Name: k}
		}
		c.settings[k] = *(v.(*Settings))
	}

	t, err := parse(v, "tunnel", func() interface{} { return &Tunnel{} })
	if err != nil {
		return nil, fmt.Errorf("parsing tunnels: %w", err)
	}
	for k, v := range t {
		if _, ok := c.tunnels[k]; ok {
			return nil, &DuplicateConfigNameError{Name: k}
		}
		c.tunnels[k] = *(v.(*Tunnel))
	}

	return &c, nil
}
