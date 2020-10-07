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
	"sort"
	"strings"

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
)

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

// Configuration loads multiple configuration files into individual viper instances and creates structs representing
// the configured hosts and credential sources.
//
// HostGlobals will always contain all possible global field names for each host. The values will be empty strings if
// a field was not included in the config file.
type Configuration struct {
	Data        map[string]*viper.Viper      // Data from individual config files
	Hosts       map[string]Host              // Host configs
	HostGlobals map[string]map[string]string // Global Host fields by [host key][field name]

	creds map[string]Cred // All unique Cred configs, by cred key name
}

// HostsSortedByPattern returns a slice of host config key strings matching the given pattern.
func (c *Configuration) HostsSortedByPattern(pattern string) []string {
	keys := c.HostKeys()
	sorter := levenshteinSort{
		keys,
		pattern,
	}

	sort.Sort(sorter)

	return sorter.items
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
		if topLevelKey != "host" && topLevelKey != "cred" {
			return fmt.Errorf("%s: user config file entry keys must start with 'host.' or 'cred.': use default"+
				" config file '%s' for all other entries", key, DefaultConfigName)
		}
	}

	return nil
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

	c.loadCredentials(c.get("cred"))
	c.loadHosts(c.get("host"))
}

// get returns the all configured items under the given key, from all config files.
func (c *Configuration) get(key string) map[string]map[string]interface{} {
	var allConfigs = make(map[string]map[string]interface{})

	for _, cfg := range c.Data {
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

			c.Hosts[itemKey] = host
			c.HostGlobals[itemKey] = make(map[string]string)

			for _, global := range GlobalHostFieldNames() {
				value, _ := d[global].(string)
				c.HostGlobals[itemKey][global] = value
			}
		}
	}
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
