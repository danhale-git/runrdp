package config

import (
	"fmt"
	"net"
	"reflect"

	"github.com/danhale-git/runrdp/internal/aws/ec2instances"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/viper"
)

// GetHost returns a host struct and it's reflect.Value based on the given type key.
func GetHost(key string) (Host, reflect.Value, error) {
	switch key {
	case "ip":
		host := IPHost{}
		return &host, reflect.ValueOf(&host).Elem(), nil
	case "awsec2":
		host := EC2Host{}
		return &host, reflect.ValueOf(&host).Elem(), nil
	default:
		return nil, reflect.Value{},
			fmt.Errorf("host type key'%s' not recognized", key)
	}
}

// IPHost defines a host to connect to using an IP or hostname.
type IPHost struct {
	Address string // IP address or hostname to connect to
	//port    int
}

// Socket returns this host's IP or hostname.
func (h *IPHost) Socket() (string, error) {
	_, err := net.LookupHost(h.Address)
	if err != nil {
		return "", fmt.Errorf("address %s is not a valid hostname or ip: %s", h.Address, err)
	}

	return h.Address, nil // :<port>
}

// Credentials returns nil as this type has no special credentials object.
func (h *IPHost) Credentials() Cred {
	return nil
}

// EC2Host defines an AWS EC2 instance to connect to by getting it's address from the AWS API.
type EC2Host struct {
	Private     bool
	GetCred     bool
	ID          string
	Profile     string
	Region      string
	IncludeTags []string
	ExcludeTags []string
	//Port    int
}

// Socket returns the public or private IP address of this instance based on the value of the Private field.
func (h *EC2Host) Socket() (string, error) {
	sess := ec2instances.NewSession(h.Profile, h.Region)

	var instance *ec2.Instance

	var err error

	switch {
	case h.ID != "":
		instance, err = ec2instances.InstanceFromID(sess, h.ID)
		if err != nil {
			return "", fmt.Errorf("getting ec2 instance by id: %s", err)
		}
	case h.IncludeTags != nil:
		filter := ec2instances.Tags{
			IncludeTags: h.IncludeTags,
			ExcludeTags: h.ExcludeTags,
			Separator:   viper.GetString("tag-separator"),
		}
		instance, err = ec2instances.InstanceFromTagFilter(sess, filter)

		if err != nil {
			return "", fmt.Errorf("getting ec2 instance by tag filter: %s", err)
		}

		h.ID = *instance.InstanceId
	default:
		return "", fmt.Errorf("no tag filters and no instance id in config, must provide one")
	}

	if *instance.State.Name != ec2.InstanceStateNameRunning {
		return "", fmt.Errorf("instance %s: state is not 'running'", *instance.InstanceId)
	}

	if h.Private {
		return *instance.PrivateIpAddress, nil
	}

	return *instance.PublicIpAddress, nil // :<port>
}

// Credentials returns the special credentials type EC2PasswordCred if the GetCred field is true, or nil if it is not.
func (h *EC2Host) Credentials() Cred {
	if h.GetCred {
		return &EC2PasswordCred{Host: h}
	}

	return nil
}
