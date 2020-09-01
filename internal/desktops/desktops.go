package desktops

import (
	"fmt"
	"reflect"

	"github.com/danhale-git/runrdp/internal/aws"
)

type Desktops map[string]Desktop

type DesktopConfig struct {
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

type IPHost struct {
	Address string
	//port    int
}

func (h IPHost) Socket() string {
	return h.Address // :<port>
}

type AWSEC2Host struct {
	ID      string
	Private bool
	Profile string
	Region  string
	//Port    int
}

func (a AWSEC2Host) Socket() string {
	session := aws.NewSession(a.Profile, a.Region)
	instance, err := aws.InstanceFromID(session, a.ID)

	if err != nil {
		fmt.Printf("error querying aws for ec2 instance: %s", err)
	}

	if a.Private {
		return *instance.PrivateIpAddress
	}

	return *instance.PublicIpAddress // :<port>
}

func LoadDesktops(desktopConfigs []DesktopConfig) Desktops {
	desktops := make(map[string]Desktop, len(desktopConfigs))

	for _, c := range desktopConfigs {
		desktop := Desktop{}

		loadHost(&desktop, c.Host.(map[interface{}]interface{}))
		desktops[desktop.Name] = desktop

		fmt.Printf("%+v\n", desktop)
	}

	return desktops
}

func loadHost(desktop *Desktop, data map[interface{}]interface{}) {
	desktop.Name = data["Name"].(string)

	switch data["Type"] {
	case "AWSEC2":
		h := AWSEC2Host{}
		fields(
			reflect.ValueOf(&h).Elem(),
			reflect.ValueOf(h).NumField(),
			data,
		)

		desktop.Host = h

	case "IP":
		h := IPHost{}
		fields(reflect.ValueOf(&h).Elem(),
			reflect.ValueOf(h).NumField(),
			data,
		)

		desktop.Host = h

	default:
		fmt.Printf("unrecognised host entry: %s\n", data["Name"])
	}
}

func fields(values reflect.Value, numFields int, data map[interface{}]interface{}) {
	types := values.Type()

	for i := 0; i < numFields; i++ {
		field := types.Field(i)
		value := values.Field(i)

		switch field.Type.Name() {
		case "bool":
			value.SetBool(data[field.Name].(bool))
		case "string":
			value.SetString(data[field.Name].(string))
		}
	}
}
