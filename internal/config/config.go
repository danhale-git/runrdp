package config

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sahilm/fuzzy"

	"github.com/danhale-git/runrdp/internal/config/hosts"

	"github.com/spf13/viper"
)

// Configuration holds data from all parsed config files as structs.
type Configuration struct {
	//Data        map[string]*viper.Viper    // Data from individual config files
	Hosts       map[string]Host              // All configured hosts
	HostGlobals map[string]map[string]string // Global Host fields by [host key][field name]. All keys exist for all hosts, undefined values are empty strings

	Creds    map[string]Cred     `mapstructure:"cred"`
	Tunnels  map[string]Tunnel   `mapstructure:"tunnel"`
	Settings map[string]Settings `mapstructure:"setting"`
}

// Host can return a hostname or IP address and/or a port.
type Host interface {
	Socket() (string, string, error)
	Validate() error
}

// Cred can return valid credentials used to authenticate an RDP session.
type Cred interface {
	Retrieve() (string, string, error)
	Validate() error
}

// ReadConfigs reads a map of io.Reader into a matching map of viper.Viper. All config files are also concatenated with
// newline delimiters and read into the global viper instance.
func ReadConfigs(readers map[string]io.Reader) (map[string]*viper.Viper, error) {
	concatConfig := make([]byte, 0)

	vipers := make(map[string]*viper.Viper)
	for k, r := range readers {
		var buf bytes.Buffer
		tee := io.TeeReader(r, &buf)

		// Read config once for individual viper instance
		v := viper.New()
		v.SetConfigType("toml")
		if err := v.ReadConfig(tee); err != nil {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		vipers[k] = v

		// Read config a second time for single concatenated viper instance
		concatConfig = append(
			concatConfig,
			append(buf.Bytes(), []byte("\n")...)...,
		)
	}

	viper.SetConfigType("toml")
	if err := viper.ReadConfig(bytes.NewReader(concatConfig)); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	return vipers, nil
}

// New takes a map of viper instances and parses them to a Configuration struct.
func New(v map[string]*viper.Viper) (*Configuration, error) {
	c := Configuration{}

	c.Hosts = make(map[string]Host)
	c.HostGlobals = make(map[string]map[string]string)
	c.Creds = make(map[string]Cred)
	c.Tunnels = make(map[string]Tunnel)
	c.Settings = make(map[string]Settings)

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
// Either a username or a password may be provided through various means. The following sources are all tried, in
// order from least to most preferred. The most preferred non-empty string is accepted for each field.
//
// - Values from calling creds.Cred.Retrieve() on the cred referred to by the global 'cred' field.
//
// - Values from calling creds.Cred.Retrieve() on the host if it implements creds.Cred.
//
// - Literal username defined in the global 'username' field. (username only)
func (c *Configuration) HostCredentials(key string) (string, string, error) {
	u := make([]string, 3)
	p := make([]string, 3)

	credKey := c.HostGlobals[key][hosts.GlobalCred.String()]

	var err error

	// Check for normal cred config entries
	if credKey != "" {
		cred, ok := c.Creds[credKey]
		if !ok {
			return "", "", fmt.Errorf("cred config '%s' not found", credKey)
		}

		u[0], p[0], err = cred.Retrieve()
		if err != nil {
			return "", "", fmt.Errorf("retrieving credentials for %s: %w", credKey, err)
		}
	}

	// Check if the host itself implements Cred
	h := c.Hosts[key]
	hostCred, ok := h.(Cred)
	if ok {
		u[1], p[1], err = hostCred.Retrieve()
	}
	if err != nil {
		return "", "", fmt.Errorf("retrieving host credentials for %s: %w", key, err)
	}

	// Check if a literal username is defined using the global username variable (not permitted for passwords)
	u[2] = c.HostGlobals[key][hosts.GlobalUsername.String()]

	// Apply the order of preference
	user, pass := lastNotEmptyStrings(u, p)

	return user, pass, nil
}

// HostSocket returns the IP/hostname and port for this host. The following sources are all tried, in
// order from least to most preferred. The most preferred non-empty string is accepted for each field.
// If noProxy is true, the config file 'proxy' global field is ignored.
//
// - Values from calling hosts.Host.Socket(), the specific behaviour of the host type.
//
// - Literal values defined in the global 'address' and 'port' fields.
//
// - Address field of the host referred to by the global 'proxy' field which points to a different host. (address only)
func (c *Configuration) HostSocket(key string, noProxy bool) (string, string, error) {
	a, p := make([]string, 3), make([]string, 3)

	var err error

	a[0], p[0], err = c.Hosts[key].Socket()
	if err != nil {
		return "", "", err
	}

	a[1], p[1] = c.HostGlobals[key][hosts.GlobalAddress.String()], c.HostGlobals[key][hosts.GlobalPort.String()]

	proxy := c.HostGlobals[key][hosts.GlobalProxy.String()]
	if proxy != "" && !noProxy {
		if c.HostExists(proxy) {
			a[2], _, err = c.HostSocket(proxy, true)
			if err != nil {
				return "", "", fmt.Errorf("retrieving socket for proxy host '%s': %s", proxy, err)
			}

		} else {
			return "", "", fmt.Errorf("proxy host '%s' does not exist", proxy)
		}
	}

	add, pass := lastNotEmptyStrings(a, p)
	return add, pass, nil
}

func lastNotEmptyStrings(a, b []string) (string, string) {
	aVal, bVal := "", ""

	for i := 0; i < 3; i++ {
		if a[i] != "" {
			aVal = a[i]
		}

		if b[i] != "" {
			bVal = b[i]
		}
	}

	return aVal, bVal
}
