package config2

import (
	"fmt"
	"io"

	"github.com/mitchellh/mapstructure"

	"github.com/danhale-git/runrdp/internal/config2/hosts"

	"github.com/spf13/viper"
)

type Configuration struct {
	//Data        map[string]*viper.Viper      // Data from individual config files
	Hosts       map[string]Host              // Host configs
	HostGlobals map[string]map[string]string // Global Host fields by [host key][field name]

	creds    map[string]Cred      `mapstructure:"cred"`
	tunnels  map[string]SSHTunnel `mapstructure:"tunnel"`
	settings map[string]Settings  `mapstructure:"setting"`
}

// Host can return a hostname or IP address and optionally a port and reference to a cred config.
type Host interface {
	Socket() (string, string, error)
}

// Cred can return valid credentials used to authenticate and RDP session.
type Cred interface {
	Retrieve() (string, string, error)
}

// SSHTunnel has the details for opening an 'SSH tunnel' (SSH port forwarding) including a reference to a Host config.
type SSHTunnel struct {
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

	c.Hosts = make(map[string]Host)
	c.HostGlobals = make(map[string]map[string]string) // TODO: parse these
	c.creds = make(map[string]Cred)                    // TODO: parse these
	c.tunnels = make(map[string]SSHTunnel)             // TODO: parse these
	c.settings = make(map[string]Settings)             // TODO: parse these

	for key, structFunc := range hosts.Map {
		if err := c.parseHosts(v, fmt.Sprintf("host.%s", key), structFunc); err != nil {
			return nil, fmt.Errorf("parsing hosts: %w", err)
		}
	}

	return &c, nil
}

func (c *Configuration) parseHosts(vipers map[string]*viper.Viper, key string, f func(int) []interface{}) error {
	for cfgName, v := range vipers {
		if !v.IsSet(key) {
			continue
		}

		all := v.Get(key).(map[string]interface{})
		structs := f(len(all))

		index := 0
		for name, raw := range all {
			if err := mapstructure.Decode(raw, &structs[index]); err != nil {
				return fmt.Errorf("decoding '%s' from config '%s': %w", key, cfgName, err)
			}

			var h Host
			h = structs[index].(Host)

			if _, ok := c.Hosts[name]; ok {
				return &DuplicateConfigNameError{Name: name}
			}

			c.Hosts[name] = h

			index++
		}
	}

	return nil
}

// DuplicateConfigNameError reports a duplicate configuration item name
type DuplicateConfigNameError struct {
	Name string
}

func (e *DuplicateConfigNameError) Error() string {
	return fmt.Sprintf("duplicate config name: '%s': all config names must be unique", e.Name)
}

// Is implements Is(error) to support errors.Is
func (e *DuplicateConfigNameError) Is(tgt error) bool {
	_, ok := tgt.(*DuplicateConfigNameError)
	if !ok {
		return false
	}
	return true
}
