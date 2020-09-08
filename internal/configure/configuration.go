package configure

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
	Files map[string]*viper.Viper
	Hosts map[string]Host
	Creds map[string]Cred
}

// Host can return a hostname or IP address and optionally a port and credential name to use.
type Host interface {
	Socket() string
}

// Cred can return valid credentials used to authenticate and RDP session.
type Cred interface {
	Retrieve() (string, string)
}

type IPHost struct {
	Address string
	//port    int
}

func (h IPHost) Socket() string {
	return h.Address // :<port>
}

func (c *Configuration) Get(key string) (interface{}, error) {
	for _, config := range c.Files {
		if config.IsSet(key) {
			return config.Get(key), nil
		}
	}

	return nil, fmt.Errorf("config entry '%s' not found", key)
}

func (c *Configuration) ReadLocalConfigFiles() {
	c.Files = make(map[string]*viper.Viper)
	c.Hosts = make(map[string]Host)
	c.Creds = make(map[string]Cred)

	c.readConfigFiles()
	c.loadConfigData()
}

// readConfigFiles reads all configuration files into viper
func (c *Configuration) readConfigFiles() {
	for _, configFile := range configFileNames() {
		c.Files[configFile] = readFile(configFile)
	}
}

// loadConfigData loads all configurations from Viper.
func (c *Configuration) loadConfigData() {
	c.Hosts = loadHosts(c.getConfig("host"))
	c.Creds = loadCredentials(c.getConfig("cred"))
}

// getConfig returns the all configurations under the given key, from all config files.
func (c *Configuration) getConfig(key string) map[string]map[string]interface{} {
	var allConfigs = make(map[string]map[string]interface{})

	for _, config := range c.Files {
		for kind, items := range config.GetStringMap(key) {
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

func loadHosts(hostsConfig map[string]map[string]interface{}) map[string]Host {
	hosts := make(map[string]Host)

	for key, data := range hostsConfig["ip"] {
		h, err := ipHost(data.(map[string]interface{}))

		if err != nil {
			fmt.Printf("Failed to load host '%s': %s\n", key, err)
			continue
		}

		hosts[key] = h
	}

	for key, data := range hostsConfig["awsec2"] {
		h, err := awsEC2Host(data.(map[string]interface{}))

		if err != nil {
			fmt.Printf("Failed to load host '%s': %s\n", key, err)
			continue
		}

		hosts[key] = h
	}

	return hosts
}

func loadCredentials(credentialsConfig map[string]map[string]interface{}) map[string]Cred {
	creds := make(map[string]Cred)

	for key, data := range credentialsConfig["awssm"] {
		c, err := secretsManager(data.(map[string]interface{}))

		if err != nil {
			fmt.Printf("Failed to load credentials '%s': %s\n", key, err)
			continue
		}

		creds[key] = c
	}

	return creds
}

func secretsManager(data map[string]interface{}) (aws.SecretsManager, error) {
	c := aws.SecretsManager{}
	err := setFields(
		reflect.ValueOf(&c).Elem(),
		reflect.ValueOf(c).NumField(),
		data,
	)

	return c, err
}

func ipHost(data map[string]interface{}) (IPHost, error) {
	h := IPHost{}
	err := setFields(
		reflect.ValueOf(&h).Elem(),
		reflect.ValueOf(h).NumField(),
		data,
	)

	return h, err
}

func awsEC2Host(data map[string]interface{}) (aws.EC2Host, error) {
	h := aws.EC2Host{}
	err := setFields(
		reflect.ValueOf(&h).Elem(),
		reflect.ValueOf(h).NumField(),
		data,
	)

	return h, err
}

// setFields uses reflection to populate the fields of a struct from values in a map. Any values not present in the map
// will be left empty in the struct.
func setFields(values reflect.Value, numFields int, data map[string]interface{} /*, optionalFields ...string*/) error {
	types := values.Type()

	for i := 0; i < numFields; i++ {
		field := types.Field(i)
		value := values.Field(i)
		fieldName := strings.ToLower(field.Name)

		d, exists := data[fieldName]
		if !exists {
			return fmt.Errorf("config key %s may be incorrect or missing", fieldName)
		}

		switch field.Type.Name() {
		case "bool":
			dt, ok := d.(bool)
			if !ok {
				return ConfigFieldLoadError{
					ConfigName: "host",
					FieldName:  fieldName,
					Message:    "expected value of type bool"}
			}

			value.SetBool(dt)

		case "string":
			dt, ok := d.(string)
			if !ok {
				return ConfigFieldLoadError{
					ConfigName: "host",
					FieldName:  fieldName,
					Message:    "expected value of type string"}
			}

			value.SetString(dt)
		}
	}

	return nil
}

type ConfigFieldLoadError struct {
	ConfigName string
	FieldName  string
	Message    string
}

func (c ConfigFieldLoadError) Error() string {
	return fmt.Sprintf("error loading '%s' field in %s config: %s", c.FieldName, c.ConfigName, c.Message)
}
