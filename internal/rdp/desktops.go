package rdp

import (
	"fmt"
	"reflect"

	"github.com/danhale-git/runrdp/internal/aws"
)

type Desktops map[string]Desktop

type DesktopConfig struct {
	Name  string
	Host  interface{}
	Creds interface{}
}

type Desktop struct {
	Name string
	Host Host
	//Creds Creds
	Port int
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

func LoadDesktops(desktopConfigs []DesktopConfig) Desktops {
	desktops := make(map[string]Desktop, len(desktopConfigs))

	for _, c := range desktopConfigs {
		desktop := Desktop{}
		desktop.Name = c.Name

		err := loadHost(&desktop, c.Host.(map[interface{}]interface{}))
		if err != nil {
			fmt.Printf("Failed to load host '%s': %s\n", c.Name, err)
			continue
		}

		desktops[desktop.Name] = desktop

		fmt.Printf("%+v\n", desktop)
	}

	return desktops
}

func loadHost(desktop *Desktop, data map[interface{}]interface{}) error {
	switch data["Type"] {
	case "AWSEC2":
		h := aws.EC2Host{}
		err := fields(
			reflect.ValueOf(&h).Elem(),
			reflect.ValueOf(h).NumField(),
			data,
		)

		if err != nil {
			return err
		}

		desktop.Host = h

	case "IP":
		h := IPHost{}
		err := fields(
			reflect.ValueOf(&h).Elem(),
			reflect.ValueOf(h).NumField(),
			data,
		)

		if err != nil {
			return err
		}

		desktop.Host = h

	default:
		return fmt.Errorf("unrecognised host type: %s", data["Type"])
	}

	return nil
}

func fields(values reflect.Value, numFields int, data map[interface{}]interface{}) error {
	types := values.Type()

	for i := 0; i < numFields; i++ {
		field := types.Field(i)
		value := values.Field(i)

		d, ok := data[field.Name]
		if !ok {
			return fmt.Errorf("config key %s may be incorrect or missing", field.Name)
		}

		switch field.Type.Name() {
		case "bool":
			dt, ok := d.(bool)
			if !ok {
				return ConfigFieldLoadError{
					ConfigName: "host",
					FieldName:  field.Name,
					Message:    "expected value of type bool"}
			}

			value.SetBool(dt)

		case "string":
			dt, ok := d.(string)
			if !ok {
				return ConfigFieldLoadError{
					ConfigName: "host",
					FieldName:  field.Name,
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
