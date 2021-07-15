package config2

import (
	"fmt"
	"io"

	"github.com/sahilm/fuzzy"

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

	if err := parseConfiguration(v, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// HostsSortedByPattern returns a slice of host config names matching the given pattern, in order closest match first.
func (c *Configuration) HostsSortedByPattern(pattern string) []string {
	keys := c.HostKeys()

	sorted := make([]string, 0)

	matches := fuzzy.Find(pattern, keys)
	for _, m := range matches {
		sorted = append(sorted, m.Str)
	}

	return sorted
}

// HostKeys returns a slice containing the names of all host config entries.
func (c *Configuration) HostKeys() []string {
	keys := make([]string, len(c.Hosts))
	index := 0

	for k := range c.Hosts {
		keys[index] = k
		index++
	}

	return keys
}

// HostExists returns true if the given host is in the configuration.
func (c *Configuration) HostExists(key string) bool {
	_, ok := c.Hosts[key]
	return ok
}

// HostCredentials returns the username and password for this host.
//
// Either a username or a password may be provided through various means. Each is checked and there is an order of
// preference to determine which is used.
//
// Sources from least to most preferred:
//
// - Config file 'cred' field credentials
//
// - Host credentials, where the Host also implements Cred
//
// - Config file global 'username' field (username only)
func (c *Configuration) HostCredentials(key string) (string, string, error) {
	username := make([]string, 3)
	password := make([]string, 3)

	credKey, ok := c.HostGlobals[key][hosts.GlobalCred.String()]

	var err error

	// Check for normal cred config entries
	if ok {
		cred, ok := c.creds[credKey]
		if !ok {
			return "", "", fmt.Errorf("cred config '%s' not found", credKey)
		}

		username[0], password[0], err = cred.Retrieve()
		if err != nil {
			return "", "", fmt.Errorf("retrieving credentials for %s: %w", credKey, err)
		}
	}

	// Check if the host itself implements creds.Cred
	h := c.Hosts[key]
	hostCred, ok := h.(creds.Cred)
	if ok {
		username[1], password[1], err = hostCred.Retrieve()
	}
	if err != nil {
		return "", "", fmt.Errorf("retrieving host credentials for %s: %w", key, err)
	}

	// Check if a literal username is defined using the global username variable (not permitted for passwords)
	username[2] = c.HostGlobals[key][hosts.GlobalUsername.String()]

	// Apply the order of preference
	user, pass := "", ""

	for i := 0; i < 3; i++ {
		if username[i] != "" {
			user = username[i]
		}

		if password[i] != "" {
			pass = password[i]
		}
	}

	return user, pass, nil
}
