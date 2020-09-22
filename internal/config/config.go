package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

const (
	// DefaultConfigName is the name of the default config file. This is the config which will be merged with the
	// given command line flags.
	DefaultConfigName = "config"
)

// Configuration load multiple configuration files into individual viper instances and creates structs representing
// the configured hosts and credential sources.
type Configuration struct {
	Data   map[string]*viper.Viper // Data from individual config files
	Hosts  map[string]Host         // Host configs
	Creds  map[string]*Cred        // Cred configs by host key name
	Proxys map[string]*Host        // Host Proxys by host key name

	creds map[string]Cred // All unique Cred configs, by cred key name
}

// Host can return a hostname or IP address and optionally a port and credential name to use.
type Host interface {
	GetAddress() (string, error) // Host.GetAddress is called first
	GetPort() int
	Credentials() Cred // Host.Credentials is called second
}

// Cred can return valid credentials used to authenticate and RDP session.
type Cred interface {
	Retrieve() (string, string, error) // Cred.Credentials is called third
}

/*// Get searches data from all config files and returns the value of the given key if it exists or an error if it
// doesn't.
func (c *Configuration) Get(key string) (interface{}, error) {
	for _, cfg := range c.Data {
		if cfg.IsSet(key) {
			return cfg.Get(key), nil
		}
	}

	return nil, fmt.Errorf("config entry '%s' not found", key)
}*/

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
	c.Creds = make(map[string]*Cred)
	c.Proxys = make(map[string]*Host)
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

			if err != nil {
				fmt.Printf("Failed to load host '%s': %s\n", itemKey, err)
				continue
			}

			err = setFields(val, data.(map[string]interface{}))

			if err != nil {
				fmt.Printf("Failed to load host '%s': %s\n", itemKey, err)
				continue
			}

			c.Hosts[itemKey] = host
			c.Creds[itemKey] = c.hostCred(data.(map[string]interface{}))
		}
	}

	// Second pass now hosts are all loaded, to get proxy hosts
	for _, item := range hostsConfig {
		for itemKey, data := range item {
			c.Proxys[itemKey] = c.hostProxy(data.(map[string]interface{}))
		}
	}
}

// hostCred returns a pointer to the Cred referenced by the cred field in this host. If no 'cred' field is defined, nil
// is silently returned (the field is optional).
func (c *Configuration) hostCred(data map[string]interface{}) *Cred {
	cKey, ok := data["cred"]
	if !ok {
		return nil
	}

	cred, ok := c.creds[cKey.(string)]
	if !ok {
		fmt.Printf("Credentials '%s' not found\n", cKey)
		return &cred
	}

	return &cred
}

// hostProxy returns a pointer to the Host referenced by the proxy field in this host config. If no 'proxy' field is
// defined, nil is silently returned (the field is optional).
func (c *Configuration) hostProxy(data map[string]interface{}) *Host {
	pKey, ok := data["proxy"]
	if !ok {
		return nil
	}

	proxy, ok := c.Hosts[pKey.(string)]
	if !ok {
		fmt.Printf("Host '%s' not found\n", pKey)
		return &proxy
	}

	return &proxy
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
	}

	for k, v := range data {
		if k == "cred" || k == "proxy" {
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
