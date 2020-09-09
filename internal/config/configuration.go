package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/danhale-git/runrdp/internal/aws"

	"github.com/spf13/viper"
)

// Configuration load multiple configuration files into individual viper instances and creates structs representing
// the configured hosts and credential sources.
type Configuration struct {
	Files map[string]*viper.Viper // Data from individual config files
	Hosts map[string]Host         // Host configs
	Creds map[string]Cred         // Cred configs by host key name
	creds map[string]Cred         // Cred configs by cred key name
}

// Host can return a hostname or IP address and optionally a port and credential name to use.
type Host interface {
	Socket() string
	Credentials() Cred
}

// Cred can return valid credentials used to authenticate and RDP session.
type Cred interface {
	Retrieve() (string, string)
}

// Get searches data from all config files and returns the value of the given key if it exists or an error if it
// doesn't.
func (c *Configuration) Get(key string) (interface{}, error) {
	for _, cfg := range c.Files {
		if cfg.IsSet(key) {
			return cfg.Get(key), nil
		}
	}

	return nil, fmt.Errorf("config entry '%s' not found", key)
}

// LoadLocalConfigFiles load host and cred configurations from all files in the default config directory.
func (c *Configuration) LoadLocalConfigFiles() {
	c.Files = make(map[string]*viper.Viper)
	c.Hosts = make(map[string]Host)
	c.Creds = make(map[string]Cred)
	c.creds = make(map[string]Cred)

	c.readConfigFiles()
	c.loadCredentials(c.getConfig("cred"))
	c.loadHosts(c.getConfig("host"))
}

// readConfigFiles reads all configuration files into viper
func (c *Configuration) readConfigFiles() {
	for _, configFile := range configFileNames() {
		c.Files[configFile] = readFile(configFile)
	}
}

// getConfig returns the all configurations under the given key, from all config files.
func (c *Configuration) getConfig(key string) map[string]map[string]interface{} {
	var allConfigs = make(map[string]map[string]interface{})

	for _, cfg := range c.Files {
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
	for key, data := range credentialsConfig["awssm"] {
		c.setCred(key, data, secretsManager)
	}
}

func (c *Configuration) loadHosts(hostsConfig map[string]map[string]interface{}) {
	for key, data := range hostsConfig["ip"] {
		c.setHost(key, data.(map[string]interface{}), ipHost)
	}

	for key, data := range hostsConfig["awsec2"] {
		c.setHost(key, data.(map[string]interface{}), awsEC2Host)
	}
}

// setCred constructs and sets a new cred config struct with the underlying value returned by the given function.
func (c *Configuration) setCred(
	key string,
	data interface{},
	f func(data map[string]interface{}) (Cred, error),
) {
	cred, err := f(data.(map[string]interface{}))

	if err != nil {
		fmt.Printf("Failed to load credentials '%s': %s\n", key, err)
	}

	c.creds[key] = cred
}

// setHost constructs and sets a new host config struct with the underlying value returned by the given function.
func (c *Configuration) setHost(
	key string, data interface{},
	f func(data map[string]interface{}) (Host, error),
) {
	h, err := f(data.(map[string]interface{}))

	if err != nil {
		fmt.Printf("Failed to load host '%s': %s\n", key, err)
	}

	c.Hosts[key] = h
	c.Creds[key] = c.getHostCred(data.(map[string]interface{}))
}

// getHostCred returns the cred struct which corresponds to the given host data. If no 'cred' field is defined, nil
// is silently returned (the field is optional). If the cred field is defined but the value is not recognised an error
// is returned.
func (c *Configuration) getHostCred(data map[string]interface{}) Cred {
	cKey, ok := data["cred"]
	if !ok {
		return nil
	}

	cred, ok := c.creds[cKey.(string)]
	if !ok {
		fmt.Printf("Credentials '%s' not found", cKey)
		return cred
	}

	return cred
}

func secretsManager(data map[string]interface{}) (Cred, error) {
	c := aws.SecretsManager{}
	err := setFields(
		reflect.ValueOf(&c).Elem(),
		reflect.ValueOf(c).NumField(),
		data,
	)

	return c, err
}

func ipHost(data map[string]interface{}) (Host, error) {
	h := IPHost{}
	err := setFields(
		reflect.ValueOf(&h).Elem(),
		reflect.ValueOf(h).NumField(),
		data,
	)

	return h, err
}

func awsEC2Host(data map[string]interface{}) (Host, error) {
	h := EC2Host{}
	err := setFields(
		reflect.ValueOf(&h).Elem(),
		reflect.ValueOf(h).NumField(),
		data,
	)

	return h, err
}

// setFields uses reflection to populate the fields of a struct from values in a map. Any values not present in the map
// will be left empty in the struct.
func setFields(values reflect.Value, numFields int, data map[string]interface{}) error {
	structType := values.Type()

	fieldMap := make(map[string]reflect.StructField)
	valueMap := make(map[string]reflect.Value)

	for i := 0; i < numFields; i++ {
		f := structType.Field(i)
		v := values.Field(i)
		fieldName := strings.ToLower(f.Name)

		fieldMap[fieldName] = f
		valueMap[fieldName] = v
	}

	for k, v := range data {
		if k == "cred" {
			continue
		}

		_, ok := fieldMap[k]
		if !ok {
			return fmt.Errorf("config key %s is invalid for type %s", k, structType.Name())
		}

		switch fieldMap[k].Type.Name() {
		case "bool":
			dt, ok := v.(bool)
			if !ok {
				return ConfigFieldLoadError{
					ConfigName: strings.ToLower(structType.Name()),
					FieldName:  k,
					Message:    "expected value of type bool"}
			}

			valueMap[k].SetBool(dt)

		case "string":
			dt, ok := v.(string)
			if !ok {
				return ConfigFieldLoadError{
					ConfigName: strings.ToLower(structType.Name()),
					FieldName:  k,
					Message:    "expected value of type string"}
			}

			valueMap[k].SetString(dt)
		}
	}

	return nil
}

// ConfigFieldLoadError reports an error loading a config field.
type ConfigFieldLoadError struct {
	ConfigName string
	FieldName  string
	Message    string
}

func (c ConfigFieldLoadError) Error() string {
	return fmt.Sprintf("error loading '%s' field in %s config: %s", c.FieldName, c.ConfigName, c.Message)
}
