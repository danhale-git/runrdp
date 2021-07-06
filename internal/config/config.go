// Package config facilitates the loading and referencing of multiple config files, each with their own viper.Viper
// instance.
/*
 */
package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sahilm/fuzzy"
	"github.com/spf13/viper"
)

// DefaultConfigName is the name of the default config file. This is the config which will be merged with the
// given command line flags.
const DefaultConfigName = "config.toml"

const (
	globalHostCred GlobalHostFields = iota
	globalHostProxy
	globalHostAddress
	globalHostPort
	globalHostUsername
	globalHostTunnel
	globalHostSettings
)

var validKeys = []string{"host", "cred", "tunnel", "settings"}

// GlobalHostFields are the names of fields which may be configured in any host.
type GlobalHostFields int

func (p GlobalHostFields) String() string {
	return GlobalHostFieldNames()[p]
}

// GlobalHostFieldNames returns a slice of field name strings corresponding to GlobalHostFields
func GlobalHostFieldNames() []string {
	return []string{
		"cred",
		"proxy",
		"address",
		"port",
		"username",
		"tunnel",
		"settings",
	}
}

// InGlobalHostFieldNames returns true if the given name is in the list of global host field names
func InGlobalHostFieldNames(name string) bool {
	for _, n := range GlobalHostFieldNames() {
		if n == name {
			return true
		}
	}

	return false
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
	Host      string // Config key for the intermediate host which will forward the connection.
	LocalPort string // Port which the intermediate host listens on.
	Key       string // Full path to the key file for
	User      string // The SSH user to connect as
}

// Settings is the configuration of .RDP file settings.
// https://docs.microsoft.com/en-us/windows-server/remote/remote-desktop-services/clients/rdp-files
type Settings struct {
	Height int // Height of the session in pixels
	Width  int // Width of the session in pixels
	Scale  int // Scale the session view (100, 125, 150, 175, 200, 250, 300, 400, 500)
}

// Configuration loads multiple configuration files into individual viper instances and creates structs representing
// the configured hosts and credential sources.
//
// HostGlobals will always contain all possible global field names for each host. The values will be empty strings if
// a field was not included in the config file.
type Configuration struct {
	Data        map[string]*viper.Viper      // Data from individual config files
	Hosts       map[string]Host              // Host configs
	HostGlobals map[string]map[string]string // Global Host fields by [host key][field name]

	creds    map[string]Cred      // All unique Cred configs, by cred key name
	tunnels  map[string]SSHTunnel // SSHTunnel configs
	settings map[string]Settings  // Settings configs
}

// HostsSortedByPattern returns a slice of host config key strings matching the given pattern.
func (c *Configuration) HostsSortedByPattern(pattern string) []string {
	keys := c.HostKeys()

	sorted := make([]string, len(keys))

	matches := fuzzy.Find(pattern, keys)
	for i, m := range matches {
		sorted[i] = m.Str
	}

	return sorted
}

// HostKeys returns a slice containing the names of all loaded host config entries from all config files.
func (c *Configuration) HostKeys() []string {
	keys := make([]string, len(c.Hosts))
	index := 0

	for k := range c.Hosts {
		keys[index] = k
		index++
	}

	return keys
}

// HostExists returns true if the given host key has been configured and successfully loaded
func (c *Configuration) HostExists(key string) bool {
	_, ok := c.Hosts[key]
	return ok
}

// HostCredentials returns the username and password for this host.
//
// Either a username or a password may be provided through various means. HostCredentials tries multiple sources in
// order from least to most preferred, overwriting where new values are found.
//
// Sources:
//
// - Config file 'cred' field credentials
//
// - Host credentials
//
// - Config file global 'username' field (username only)
func (c *Configuration) HostCredentials(key string) (string, string, error) {
	var username, password string

	h := c.Hosts[key]

	credKey := c.HostGlobals[key][globalHostCred.String()]

	if credKey != "" {
		cred, ok := c.creds[credKey]
		if !ok {
			return "", "", fmt.Errorf("cred config '%s' not found", credKey)
		}

		err := overwriteValues(&username, &password, cred.Retrieve)

		if err != nil {
			return "", "", err
		}
	}

	hostCred, ok := h.(Cred)

	if ok {
		err := overwriteValues(&username, &password, hostCred.Retrieve)

		if err != nil {
			return "", "", err
		}
	}

	uGlobal := c.HostGlobals[key][globalHostUsername.String()]

	if uGlobal != "" {
		username = uGlobal
	}

	return username, password, nil
}

// HostSocket returns the IP/hostname and port for this host.
//
// Either a hostname or a port may be provided through various means. HostSocket tries multiple sources in
// order from least to most preferred, overwriting where new values are found.
//
// If noProxy is true, the config file 'proxy' global field is ignored.
//
// Sources:
//
// - Host socket
//
// - Config file global 'address' and 'port' fields
//
// - Config file global 'proxy' field socket address (address only)
func (c *Configuration) HostSocket(key string, noProxy bool) (string, string, error) {
	h := c.Hosts[key]

	address, port, err := h.Socket()
	if err != nil {
		return "", "", err
	}

	_ = overwriteValues(
		&address,
		&port,
		func() (string, string, error) {
			return c.HostGlobals[key][globalHostAddress.String()], c.HostGlobals[key][globalHostPort.String()], nil
		},
	)

	proxyKey := c.HostGlobals[key][globalHostProxy.String()]
	if proxyKey != "" && !noProxy {
		if c.HostExists(proxyKey) {
			aProxy, _, err := c.HostSocket(proxyKey, true)

			if err != nil {
				return "", "", fmt.Errorf("retrieving socket for proxy host '%s': %s", proxyKey, err)
			}

			if aProxy != "" {
				address = aProxy
			}
		} else {
			return "", "", fmt.Errorf("proxy host '%s' does not exist", proxyKey)
		}
	}

	return address, port, nil
}

// HostTunnel returns a pointer to the SSH tunnel associated with this host, if there is one. If no tunnel is configured
// the pointer will be nil.
func (c *Configuration) HostTunnel(key string) (*SSHTunnel, error) {
	k, ok := c.HostGlobals[key][globalHostTunnel.String()]

	if !ok || k == "" {
		return nil, nil
	}

	t, ok := c.tunnels[k]

	if !ok {
		return nil, fmt.Errorf("host '%s' references ssh tunnel '%s' which does not appear in config", key, k)
	}

	return &t, nil
}

// HostSettings returns a pointer to the Settings object associated with the given host. If no Settings exist for this
// host, the return value will be nil.
func (c *Configuration) HostSettings(key string) (*Settings, error) {
	k, ok := c.HostGlobals[key][globalHostSettings.String()]

	if !ok || k == "" {
		return nil, nil
	}

	s, ok := c.settings[k]

	if !ok {
		return nil, fmt.Errorf("host '%s' references settings '%s' which does not appear in config", key, k)
	}

	return &s, nil
}

// Assign the results of newValues to a and b respectively, unless they are empty strings.
func overwriteValues(a, b *string, newValues func() (string, string, error)) error {
	aNew, bNew, err := newValues()

	if err != nil {
		return err
	}

	if aNew != "" {
		*a = aNew
	}

	if bNew != "" {
		*b = bNew
	}

	return nil
}

// ReadConfigFiles reads each file in the config directory as an individual Viper instance.
func (c *Configuration) ReadConfigFiles() {
	c.Data = make(map[string]*viper.Viper)

	for _, fileName := range configFileNames() {
		newConfig := readFile(fileName)
		err := validateConfig(newConfig)

		if err != nil {
			fmt.Printf("Config file '%s' is invalid: %s\n", fileName, err)
		} else {
			c.Data[fileName] = newConfig
		}
	}
}

func validateConfig(v *viper.Viper) error {
	for _, key := range v.AllKeys() {
		topLevelKey := strings.Split(key, ".")[0]
		if !keyIsValid(topLevelKey) {
			return fmt.Errorf("%s: user config file entry keys must start with 'host.' or 'cred.': use default"+
				" config file '%s' for all other entries", key, DefaultConfigName)
		}
	}

	return nil
}

func keyIsValid(key string) bool {
	for _, v := range validKeys {
		if key == v {
			return true
		}
	}

	return false
}

func configFileNames() []string {
	files, err := ioutil.ReadDir(viper.GetString("config-root"))
	if err != nil {
		return []string{}
	}

	names := make([]string, 0)

	for _, f := range files {
		n := f.Name()
		if strings.TrimSpace(n) == "" {
			continue
		}

		if n == DefaultConfigName {
			continue
		}

		names = append(names, n)
	}

	return names
}

func readFile(name string) *viper.Viper {
	newViper := viper.New()

	newViper.SetConfigType("toml")
	newViper.SetConfigFile(filepath.Join(
		viper.GetString("config-root"),
		name,
	))

	if err := newViper.ReadInConfig(); err != nil {
		fmt.Printf("failed to load %s config: %v\n", name, err)
	}

	return newViper
}

// BuildData constructs Host and Cred objects from all available config data.
func (c *Configuration) BuildData() {
	c.Hosts = make(map[string]Host)
	c.HostGlobals = make(map[string]map[string]string)
	c.creds = make(map[string]Cred)
	c.tunnels = make(map[string]SSHTunnel)
	c.settings = make(map[string]Settings)

	c.loadCredentials(c.getNested("cred"))
	c.loadHosts(c.getNested("host"))
	c.loadTunnels(c.get("tunnel"))
	c.loadSettings(c.get("settings"))
}

// getNested returns the all configured items under the given key, where the item key has 3 labels.
func (c *Configuration) getNested(key string) map[string]map[string]interface{} {
	var allConfigs = make(map[string]map[string]interface{})

	// Iterate config files
	for _, cfg := range c.Data {
		// Iterate items with key in config file
		for kind, items := range cfg.GetStringMap(key) {
			if _, ok := allConfigs[kind]; !ok {
				allConfigs[kind] = make(map[string]interface{})
			}

			for k, v := range items.(map[string]interface{}) {
				allConfigs[kind][k] = v
			}
		}
	}

	return allConfigs
}

// get returns the all configured items under the given key, where the item key has 2 labels.
func (c *Configuration) get(key string) map[string]interface{} {
	var allConfigs = make(map[string]interface{})

	// Iterate config files
	for _, cfg := range c.Data {
		for k, v := range cfg.GetStringMap(key) {
			allConfigs[k] = v
		}
	}

	return allConfigs
}

func (c *Configuration) loadTunnels(tunnelsConfig map[string]interface{}) {
	for itemKey, data := range tunnelsConfig {
		tunnel := SSHTunnel{}
		val := reflect.ValueOf(&tunnel).Elem()

		err := setFields(val, data.(map[string]interface{}))

		if err != nil {
			fmt.Printf("Failed to load tunnel '%s': %s\n", itemKey, err)
			continue
		}

		c.tunnels[itemKey] = tunnel
	}
}

func (c *Configuration) loadSettings(settingsConfig map[string]interface{}) {
	for itemKey, data := range settingsConfig {
		settings := Settings{}
		val := reflect.ValueOf(&settings).Elem()

		err := setFields(val, data.(map[string]interface{}))

		if err != nil {
			fmt.Printf("Failed to load settings '%s': %s\n", itemKey, err)
			continue
		}

		if settings.Width != 0 && (settings.Width < 200 || settings.Width > 8192) {
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
		}

		c.settings[itemKey] = settings
	}
}

func (c *Configuration) loadCredentials(credentialsConfig map[string]map[string]interface{}) {
	for typeKey, item := range credentialsConfig {
		for itemKey, data := range item {
			cred, val, err := GetCredential(typeKey)

			if err != nil {
				fmt.Printf("Failed to load credentials '%s': %s\n", itemKey, err)
				continue
			}

			err = setFields(val, data.(map[string]interface{}))

			if err != nil {
				fmt.Printf("Failed to load credentials '%s': %s\n", itemKey, err)
				continue
			}

			c.creds[itemKey] = cred
		}
	}
}

func (c *Configuration) loadHosts(hostsConfig map[string]map[string]interface{}) {
	for typeKey, item := range hostsConfig {
		for itemKey, data := range item {
			host, val, err := GetHost(typeKey)

			d := data.(map[string]interface{})

			if err != nil {
				fmt.Printf("Failed to load host '%s': %s\n", itemKey, err)
				continue
			}

			err = setFields(val, d)

			if err != nil {
				fmt.Printf("Failed to load host '%s': %s\n", itemKey, err)
				continue
			}

			globals, err := getGlobals(d)

			if err != nil {
				fmt.Printf("Failed to load host '%s': %s\n", itemKey, err)
				continue
			}

			c.Hosts[itemKey] = host
			c.HostGlobals[itemKey] = globals
		}
	}
}

func getGlobals(data map[string]interface{}) (map[string]string, error) {
	globals := make(map[string]string)

	for _, global := range GlobalHostFieldNames() {
		value, ok := data[global].(string)

		if data[global] != nil && !ok {
			return nil, fmt.Errorf("global field '%s' must be a string", global)
		}

		globals[global] = value
	}

	return globals, nil
}

// setFields uses reflection to populate the fields of a struct from values in a map. Any values not present in the map
// will be left empty in the struct.
func setFields(values reflect.Value, data map[string]interface{}) error {
	structType := values.Type()

	valueMap := make(map[string]reflect.Value)

	for i := 0; i < structType.NumField(); i++ {
		v := values.Field(i)
		fieldName := strings.ToLower(structType.Field(i).Name)

		valueMap[fieldName] = v

		if InGlobalHostFieldNames(fieldName) {
			panic(fmt.Sprintf("config type '%s' contains field '%s' which is a global host field name",
				structType.Name(), fieldName))
		}
	}

	for k, v := range data {
		if InGlobalHostFieldNames(k) {
			continue
		}

		_, exists := valueMap[k]
		if !exists {
			return fmt.Errorf("config key %s is invalid for type %s", k, structType.Name())
		}

		value := valueMap[k]
		n := strings.ToLower(structType.Name())

		switch value.Kind() {
		case reflect.Bool:
			dt, ok := v.(bool)
			if !ok {
				return FieldLoadError{ConfigName: n, FieldName: k,
					Message: "expected value of type bool"}
			}

			value.SetBool(dt)

		case reflect.Int:
			dt, ok := v.(int64)
			if !ok {
				return FieldLoadError{ConfigName: n, FieldName: k,
					Message: "expected value of type integer"}
			}

			value.SetInt(dt)

		case reflect.String:
			dt, ok := v.(string)
			if !ok {
				return FieldLoadError{ConfigName: n, FieldName: k,
					Message: "expected value of type string"}
			}

			value.SetString(dt)

		case reflect.Map:
			dt, ok := v.(map[string]interface{})
			if !ok {
				return FieldLoadError{ConfigName: n, FieldName: k,
					Message: "expected value map[string]interface{} ({ key1 = \"val1\", key2 = \"val2\" })"}
			}

			if value.IsNil() {
				value.Set(reflect.MakeMap(value.Type()))
			}

			for key, val := range dt {
				kVal := reflect.ValueOf(key)
				vVal := reflect.ValueOf(val)
				value.SetMapIndex(kVal, vVal)
			}

		case reflect.Slice:
			dt, ok := v.([]interface{})
			if !ok {
				return FieldLoadError{ConfigName: n, FieldName: k,
					Message: "expected value of type array"}
			}

			if value.IsNil() {
				value.Set(reflect.MakeSlice(value.Type(), len(dt), cap(dt)))
			}

			for i, item := range dt {
				val, ok := item.(string)
				if !ok {
					return FieldLoadError{ConfigName: n, FieldName: k,
						Message: fmt.Sprintf(`array item %d: expected value of type string (["Key1:Val1", "Key2:Val2", "KeyOnly"])`, i)}
				}

				value.Index(i).Set(reflect.ValueOf(val))
			}
		}
	}

	return nil
}

// FieldLoadError reports an error loading a config field.
type FieldLoadError struct {
	ConfigName string
	FieldName  string
	Message    string
}

func (c FieldLoadError) Error() string {
	return fmt.Sprintf("error loading '%s' field in %s config: %s", c.FieldName, c.ConfigName, c.Message)
}
