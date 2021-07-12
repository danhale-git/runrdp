package config2

import (
	"fmt"
	"io"
	"reflect"
	"strings"

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
			h := structs[index].(Host)
			value := reflect.ValueOf(h).Elem()
			if err := setFields(value, raw.(map[string]interface{})); err != nil {
				return fmt.Errorf("reading '%s' fields for %s in '%s': %w", key, name, cfgName, err)
			}

			if _, ok := c.Hosts[name]; ok {
				return &DuplicateConfigNameError{Name: name}
			}

			c.Hosts[name] = h

			index++
		}
	}

	return nil
}

// setFields uses reflection to populate the fields of a struct from values in a map. Any values not present in the map
// will be left empty in the struct.
func setFields(values reflect.Value, data map[string]interface{}) error {
	structType := values.Type()

	// Map fields to their lower case names
	valueMap := make(map[string]reflect.Value)
	for i := 0; i < structType.NumField(); i++ {
		v := values.Field(i)
		fieldName := strings.ToLower(structType.Field(i).Name)

		valueMap[fieldName] = v

		if hosts.FieldNameIsGlobal(fieldName) {
			panic(fmt.Sprintf("config type '%s' contains field '%s' which is a global host field name",
				structType.Name(), fieldName))
		}
	}

	// Iterate over all the values given in the config entry
	for k, v := range data {
		if hosts.FieldNameIsGlobal(k) {
			continue
		}

		// Check if the config entry has a corresponding field in the struct
		_, exists := valueMap[k]
		if !exists {
			return fmt.Errorf("config key %s is invalid for type %s", k, structType.Name())
		}

		value := valueMap[k]
		n := strings.ToLower(structType.Name())

		// Validate the config value and assign it to the struct field
		switch value.Kind() {
		case reflect.Bool:
			// TODO: an integer will fail this but a string will not. This should be checked against true/fale/True/False
			dt, ok := v.(bool)
			if !ok {
				return &FieldLoadError{ConfigName: n, FieldName: k,
					Message: "expected value of type bool"}
			}

			value.SetBool(dt)

		case reflect.Int:
			dt, ok := v.(int64)
			if !ok {
				return &FieldLoadError{ConfigName: n, FieldName: k,
					Message: "expected value of type integer"}
			}

			value.SetInt(dt)

		case reflect.String:
			dt, ok := v.(string)
			if !ok {
				return &FieldLoadError{ConfigName: n, FieldName: k,
					Message: "expected value of type string"}
			}

			value.SetString(dt)

		case reflect.Map:
			dt, ok := v.(map[string]interface{})
			if !ok {
				return &FieldLoadError{ConfigName: n, FieldName: k,
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
				return &FieldLoadError{ConfigName: n, FieldName: k,
					Message: "expected value of type array"}
			}

			if value.IsNil() {
				value.Set(reflect.MakeSlice(value.Type(), len(dt), cap(dt)))
			}

			for i, item := range dt {
				val, ok := item.(string)
				if !ok {
					return &FieldLoadError{ConfigName: n, FieldName: k,
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

func (e *FieldLoadError) Error() string {
	return fmt.Sprintf("error loading '%s' field in %s config: %s", e.FieldName, e.ConfigName, e.Message)
}

// Is implements Is(error) to support errors.Is
func (e *FieldLoadError) Is(tgt error) bool {
	_, ok := tgt.(*FieldLoadError)
	if !ok {
		return false
	}
	return true
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
