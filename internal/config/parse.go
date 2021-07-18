package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/danhale-git/runrdp/internal/config/creds"

	"github.com/danhale-git/runrdp/internal/config/hosts"
	"github.com/spf13/viper"
)

func parseConfiguration(v map[string]*viper.Viper, c *Configuration) error {
	if err := parseHosts(v, c.Hosts, c.HostGlobals); err != nil {
		return fmt.Errorf("parsing hosts: %w", err)
	}

	if err := parseCreds(v, c.Creds); err != nil {
		return fmt.Errorf("parsing creds: %w", err)
	}

	if err := parseSettings(v, c.Settings); err != nil {
		return fmt.Errorf("parsing settings: %w", err)
	}

	if err := parseTunnels(v, c.Tunnels); err != nil {
		return fmt.Errorf("parsing tunnels: %w", err)
	}

	return nil
}

func parseHosts(v map[string]*viper.Viper, hm map[string]Host, gm map[string]map[string]string) error {
	for key, typeFunc := range hosts.Map {
		h, err := parse(v, fmt.Sprintf("host.%s", key), typeFunc)
		if err != nil {
			return err
		}

		for k, v := range h {
			if _, ok := hm[k]; ok {
				return &DuplicateConfigNameError{Name: k}
			}
			hm[k] = v.(Host)

			if err := hm[k].Validate(); err != nil {
				return &InvalidConfigError{Reason: fmt.Errorf("%s configuration is invalid: %w", k, err)}
			}
		}

		g, err := parseGlobals(v, fmt.Sprintf("host.%s", key))
		for k, v := range g {
			gm[k] = v.(map[string]string)
		}
	}

	return nil
}

func parseCreds(v map[string]*viper.Viper, m map[string]Cred) error {
	for key, typeFunc := range creds.Map {
		cr, err := parse(v, fmt.Sprintf("cred.%s", key), typeFunc)
		if err != nil {
			return err
		}

		for k, v := range cr {
			if _, ok := m[k]; ok {
				return &DuplicateConfigNameError{Name: k}
			}
			m[k] = v.(Cred)

			if err := m[k].Validate(); err != nil {
				return &InvalidConfigError{Reason: fmt.Errorf("%s configuration is invalid: %w", k, err)}
			}
		}
	}

	return nil
}

func parseSettings(v map[string]*viper.Viper, m map[string]Settings) error {
	s, err := parse(v, "settings", func() interface{} { return &Settings{} })
	if err != nil {
		return err
	}

	for k, v := range s {
		if _, ok := m[k]; ok {
			return &DuplicateConfigNameError{Name: k}
		}
		m[k] = *(v.(*Settings))

		if err := m[k].Validate(); err != nil {
			return &InvalidConfigError{Reason: fmt.Errorf("%s configuration is invalid: %w", k, err)}
		}
	}

	return nil
}

func parseTunnels(v map[string]*viper.Viper, m map[string]Tunnel) error {
	t, err := parse(v, "tunnel", func() interface{} { return &Tunnel{} })
	if err != nil {
		return err
	}

	for k, v := range t {
		if _, ok := m[k]; ok {
			return &DuplicateConfigNameError{Name: k}
		}
		m[k] = *(v.(*Tunnel))

		if err := m[k].Validate(); err != nil {
			return &InvalidConfigError{Reason: fmt.Errorf("%s configuration is invalid: %w", k, err)}
		}
	}

	return nil
}

func parse(vipers map[string]*viper.Viper, key string, typeFunc func() interface{}) (map[string]interface{}, error) {
	parsed := make(map[string]interface{})

	for cfgName, v := range vipers {
		if !v.IsSet(key) {
			continue
		}

		all := v.Get(key).(map[string]interface{})

		for name, raw := range all {
			h := typeFunc()
			value := reflect.ValueOf(h).Elem()
			if err := setFields(value, raw.(map[string]interface{})); err != nil {
				return nil, fmt.Errorf("reading '%s' fields for %s in '%s': %w", key, name, cfgName, err)
			}

			parsed[name] = h
		}
	}

	return parsed, nil
}

func parseGlobals(vipers map[string]*viper.Viper, key string) (map[string]interface{}, error) {
	parsed := make(map[string]interface{})

	for _, v := range vipers {
		if !v.IsSet(key) {
			continue
		}

		all := v.Get(key).(map[string]interface{})

		for name, raw := range all {
			g, err := getGlobals(raw.(map[string]interface{}))
			if err != nil {
				return nil, err
			}

			parsed[name] = g
		}
	}

	return parsed, nil
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
						Message: fmt.Sprintf(`array item %d: expected value of type string (["a", "b", "c"])`, i)}
				}

				value.Index(i).Set(reflect.ValueOf(val))
			}
		}
	}

	return nil
}

func getGlobals(data map[string]interface{}) (map[string]string, error) {
	globals := make(map[string]string)

	for _, global := range hosts.GlobalFieldNames() {
		value, ok := data[global].(string)

		if data[global] != nil && !ok {
			return nil, fmt.Errorf("global field '%s' must be a string", global)
		}

		globals[global] = value
	}

	return globals, nil
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

// InvalidConfigError reports a configuration which was parsed successfully but has invalid values
type InvalidConfigError struct {
	Reason error
}

func (e *InvalidConfigError) Error() string {
	return fmt.Sprintf("config validation failed: %s", e.Reason)
}

// Is implements Is(error) to support errors.Is
func (e *InvalidConfigError) Is(tgt error) bool {
	_, ok := tgt.(*InvalidConfigError)
	if !ok {
		return false
	}
	return true
}
