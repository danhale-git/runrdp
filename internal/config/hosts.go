package config

import (
	"fmt"
	"net"

	"github.com/danhale-git/runrdp/internal/aws"
)

// IPHost defines a host to connect to using an IP or hostname.
type IPHost struct {
	Address string // IP address or hostname to connect to
	//port    int
}

// Socket returns this host's IP or hostname.
func (h IPHost) Socket() (string, error) {
	_, err := net.LookupHost(h.Address)
	if err != nil {
		return "", fmt.Errorf("address %s is not a valid hostname or ip: %s", h.Address, err)
	}
	return h.Address, nil // :<port>
}

// Credentials returns nil as this type has no special credentials object.
func (h IPHost) Credentials() Cred {
	return nil
}

// EC2Host defines an AWS EC2 instance to connect to by getting it's address from the AWS API.
type EC2Host struct {
	Private bool
	GetCred bool
	ID      string
	Profile string
	Region  string
	//Port    int
}

// Socket returns the public or private IP address of this instance based on the value of the Private field.
func (h EC2Host) Socket() (string, error) {
	sess := aws.NewSession(h.Profile, h.Region)
	instance, err := aws.InstanceFromID(sess, h.ID)

	if err != nil {
		return "", fmt.Errorf("getting ec2 instance: %s", err)
	}

	if h.Private {
		return *instance.PrivateIpAddress, nil
	}

	return *instance.PublicIpAddress, nil // :<port>
}

// Credentials returns the special credentials type EC2GetPassword if the GetCred field is true, or nil if it is not.
func (h EC2Host) Credentials() Cred {
	if h.GetCred {
		return EC2GetPassword{Host: &h}
	}

	return nil
}
