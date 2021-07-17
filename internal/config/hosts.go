package config

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/spf13/viper"

	"github.com/danhale-git/runrdp/internal/aws/ec2instances"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// GetHost returns a host struct and it's reflect.Value based on the given type key.
func GetHost(key string) (Host, reflect.Value, error) {
	switch key {
	case "basic":
		host := BasicHost{}
		return &host, reflect.ValueOf(&host).Elem(), nil
	case "awsec2":
		host := EC2Host{}
		return &host, reflect.ValueOf(&host).Elem(), nil
	default:
		return nil, reflect.Value{},
			fmt.Errorf("host type key'%s' not recognized", key)
	}
}

// BasicHost defines a host to connect to using an IP or hostname.
//
// Users may configure a basic host with global fields only (see config.go). No other fields are defined.
type BasicHost struct{}

// Socket returns this host's IP or hostname.
func (h *BasicHost) Socket() (string, string, error) {
	return "", "", nil
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

	svc ec2iface.EC2API
}

// Socket returns the public or private IP address of this instance based on the value of the Private field.
func (h *EC2Host) Socket() (string, string, error) {
	instance, err := h.Instance()

	if err != nil {
		return "", "", fmt.Errorf("getting ec2 instance: %s", err)
	}

	if *instance.State.Name != ec2.InstanceStateNameRunning {
		return "", "", fmt.Errorf("instance %s: state is not 'running'", *instance.InstanceId)
	}

	if h.Private {
		return *instance.PrivateIpAddress, "", nil
	}

	return *instance.PublicIpAddress, "", nil
}

// Instance returns the EC2 instance defined by this host. The Host ID field will be set if it is empty, so future calls
// will use the ID instead of tag filters to identify the instance.
func (h *EC2Host) Instance() (*ec2.Instance, error) {
	if h.svc == nil {
		h.svc = ec2instances.NewSession(h.Profile, h.Region)
	}

	instance, err := ec2instances.GetInstance(
		h.svc,
		h.ID,
		viper.GetString("tag-separator"),
		h.IncludeTags,
		h.ExcludeTags,
	)

	if err != nil {
		return nil, fmt.Errorf("looking for instance in region '%s': %w", h.Region, err)
	}

	if h.ID == "" {
		h.ID = *instance.InstanceId
	}

	return instance, nil
}
