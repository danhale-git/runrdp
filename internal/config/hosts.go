package config

import (
	"fmt"
	"net"
	"reflect"

	"github.com/spf13/viper"

	"github.com/danhale-git/runrdp/internal/aws/ec2instances"

	"github.com/aws/aws-sdk-go/service/ec2"
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
	Port    int
}

// GetAddress returns this host's IP or hostname.
func (h *IPHost) GetAddress() (string, error) {
	_, err := net.LookupHost(h.Address)
	if err != nil {
		return "", fmt.Errorf("address %s is not a valid hostname or ip: %s", h.Address, err)
	}

	return h.Address, nil // :<port>
}

// GetPort returns the host's configured port or an empty string if it's value is 0.
func (h *IPHost) GetPort() int {
	return h.Port
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
	Port        int
}

// GetAddress returns the public or private IP address of this instance based on the value of the Private field.
func (h *EC2Host) GetAddress() (string, error) {
	svc := ec2instances.NewSession(h.Profile, h.Region)

	instance, err := ec2instances.GetInstance(
		svc,
		h.ID,
		viper.GetString("tag-separator"),
		h.IncludeTags,
		h.ExcludeTags,
	)

	if err != nil {
		return "", fmt.Errorf("getting ec2 instance: %s", err)
	}

	if *instance.State.Name != ec2.InstanceStateNameRunning {
		return "", fmt.Errorf("instance %s: state is not 'running'", *instance.InstanceId)
	}

	if h.Private {
		return *instance.PrivateIpAddress, nil
	}

	return *instance.PublicIpAddress, nil // :<port>
}

// GetPort returns the host's configured port or an empty string if it's value is 0.
func (h *EC2Host) GetPort() int {
	return h.Port
}

// Credentials returns the special credentials type EC2PasswordCred if the GetCred field is true, or nil if it is not.
func (h *EC2Host) Credentials() Cred {
	if h.GetCred {
		return &EC2PasswordCred{Host: h}
	}

	return nil
}
