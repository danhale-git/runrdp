package configure

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/danhale-git/runrdp/internal/aws"
)

type DesktopConfig struct {
	Host        interface{}
	Credentials interface{}
}

type Desktop struct {
	Host        Host
	Credentials Credentials
	Port        int
}

type Host interface {
	Socket() string
}

type Credentials interface {
	Retrieve() (string, string)
}

type IPHost struct {
	Address string
	//port    int
}

func (h IPHost) Socket() string {
	return h.Address // :<port>
}

// LoadDesktopConfigurations converts the configuration data into structs
// which are used when starting RDP sessions.
func LoadDesktopConfigurations(creds, hosts map[string]map[string]interface{}) map[string]Desktop {
	desktops := make(map[string]Desktop, len(hosts))

	allCredentials := make(map[string]Credentials)

	for key, data := range creds["awssm"] {
		c, err := secretsManager(data.(map[string]interface{}))

		if err != nil {
			fmt.Printf("Failed to load credentials '%s': %s\n", key, err)
			continue
		}

		allCredentials[key] = c
	}

	for key, data := range hosts["ip"] {
		h, cred, err := ipHost(data.(map[string]interface{}))

		if err != nil {
			fmt.Printf("Failed to load host '%s': %s\n", key, err)
			continue
		}

		d, err := newDesktop(h, cred, allCredentials)

		if err != nil {
			fmt.Printf("Failed to load host '%s': %s\n", key, err)
			continue
		}

		desktops[key] = d
	}

	for key, data := range hosts["awsec2"] {
		h, cred, err := awsEC2Host(data.(map[string]interface{}))

		if err != nil {
			fmt.Printf("Failed to load host '%s': %s\n", key, err)
			continue
		}

		d, err := newDesktop(h, cred, allCredentials)

		if err != nil {
			fmt.Printf("Failed to load host '%s': %s\n", key, err)
			continue
		}

		if cred == "" {
			d.Credentials = aws.EC2GetPassword{EC2Host: &h}
		}

		desktops[key] = d
	}

	return desktops
}

func newDesktop(host Host, cred string, allCreds map[string]Credentials) (Desktop, error) {
	if cred == "" {
		return Desktop{Host: host}, nil
	}

	c, ok := allCreds[cred]
	if ok {
		return Desktop{Host: host, Credentials: c}, nil
	}

	return Desktop{}, fmt.Errorf("credential '%s' not found", cred)
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

func ipHost(data map[string]interface{}) (IPHost, string, error) {
	h := IPHost{}
	err := setFields(
		reflect.ValueOf(&h).Elem(),
		reflect.ValueOf(h).NumField(),
		data,
	)

	return h, data["credentials"].(string), err
}

func awsEC2Host(data map[string]interface{}) (aws.EC2Host, string, error) {
	h := aws.EC2Host{}
	err := setFields(
		reflect.ValueOf(&h).Elem(),
		reflect.ValueOf(h).NumField(),
		data,
	)

	return h, data["credentials"].(string), err
}

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
