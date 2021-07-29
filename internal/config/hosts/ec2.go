package hosts

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/danhale-git/runrdp/internal/config/hosts/ec2"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// EC2Struct returns a struct of type hosts.EC2.
func EC2Struct() interface{} {
	return &EC2{}
}

// Validate returns an error if a config field is invalid.
func (e EC2) Validate() error {
	return nil
}

// EC2 defines an AWS EC2 instance to connect to by getting it's address from the AWS API.
type EC2 struct {
	Private bool
	GetCred bool
	ID      string
	Profile string
	Region  string

	svc ec2iface.EC2API

	FilterJSON string

	fetched             bool // True if id, keyName, name, publicIP and privateIP have been fetched from the API
	id, keyName, name   *string
	publicIP, privateIP *string
}

func (e *EC2) fetch() error {
	if e.fetched {
		return nil
	}
	e.svc = ec2.NewSession(e.Profile, e.Region)

	instances, err := ec2.GetInstances(e.svc, e.ID, e.FilterJSON)
	if err != nil {
		return fmt.Errorf("getting instances: %w", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("no instances found")
	}

	instance := instances[0]

	if len(instances) > 1 {
		instance, err = ec2.ChooseInstance(instances, os.Stdin)
		if err != nil {
			return err
		}
	}

	if !ec2.IsRunning(instance) {
		return fmt.Errorf("instance state is not 'running'")
	}

	e.name = ec2.GetTag(instance, "Name")
	e.id = instance.InstanceId
	e.keyName = instance.KeyName
	e.publicIP, e.privateIP = instance.PublicIpAddress, instance.PrivateIpAddress

	e.fetched = true

	return nil
}

// Socket returns the public or private IP address of this instance based on the value of the Private field.
func (e *EC2) Socket() (string, string, error) {
	if err := e.fetch(); err != nil {
		return "", "", fmt.Errorf("fetching instance details: %w", err)
	}

	if e.Private {
		if e.privateIP == nil {
			return "", "", fmt.Errorf("instance does not have a private ip address")
		}
		return *e.privateIP, "", nil
	}

	if e.publicIP == nil {
		return "", "", fmt.Errorf("instance does not have a public ip address")
	}
	return *e.publicIP, "", nil
}

// Retrieve returns the administrator credentials for this instance or an error if unable to retrieve them. If the
// getcred field is not set it returns empty strings and no error.
func (e *EC2) Retrieve() (string, string, error) {
	if !e.GetCred {
		return "", "", nil
	}

	keyData, err := ec2.ReadPrivateKey(
		viper.GetString("ssh-directory"),
		*e.keyName,
	)

	if err != nil {
		return "", "", fmt.Errorf("reading private key: %s", err)
	}

	password, err := ec2.GetPassword(
		e.svc,
		*e.id,
		keyData,
	)
	if err != nil {
		return "", "", fmt.Errorf("getting ec2 administrator credentials: %s", err)
	}

	username := "Administrator"

	return username, password, nil
}
