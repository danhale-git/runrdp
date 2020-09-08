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

func (c *Configuration) loadCredentials(credentialsConfig map[string]map[string]interface{}) {
	for key, data := range credentialsConfig["awssm"] {
		c.setCred(key, data, secretsManager)
	}
}

func (c *Configuration) loadHosts(hostsConfig map[string]map[string]interface{}) {
	for key, data := range hostsConfig["ip"] {
		c.setHost(key, data.(map[string]interface{}), ipHost)

		c.Creds[key] = c.getHostCred(data)
	}

	for key, data := range hostsConfig["awsec2"] {
		c.setHost(key, data.(map[string]interface{}), awsEC2Host)

		c.Creds[key] = c.getHostCred(data)
	}
}

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

func (c *Configuration) setHost(
	key string,
	data interface{},
	f func(data map[string]interface{}) (Host, error),
) {
	h, err := f(data.(map[string]interface{}))

	if err != nil {
		fmt.Printf("Failed to load host '%s': %s\n", key, err)
	}

	c.Hosts[key] = h
}

func (c *Configuration) getHostCred(data interface{}) Cred {
	cKey, ok := data.(map[string]interface{})["cred"]
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
					ConfigName: "host",
					FieldName:  k,
					Message:    "expected value of type bool"}
			}

			valueMap[k].SetBool(dt)

		case "string":
			dt, ok := v.(string)
			if !ok {
				return ConfigFieldLoadError{
					ConfigName: "host",
					FieldName:  k,
					Message:    "expected value of type string"}
			}

			valueMap[k].SetString(dt)
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
