package hosts

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/danhale-git/runrdp/internal/aws/ec2instances"
	"github.com/spf13/viper"
)

// EC2TagSeparator is the delimeter use when defining array values on the command line.
const EC2TagSeparator = ";"

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
func (e *EC2) Socket() (string, string, error) {
	instance, err := e.Instance()

	if err != nil {
		return "", "", fmt.Errorf("getting ec2 instance: %s", err)
	}

	if *instance.State.Name != ec2.InstanceStateNameRunning {
		return "", "", fmt.Errorf("instance %s: state is not 'running'", *instance.InstanceId)
	}

	if e.Private {
		return *instance.PrivateIpAddress, "", nil
	}

	return *instance.PublicIpAddress, "", nil
}

// Instance returns the EC2 instance defined by this host. The Host ID field will be set if it is empty, so future calls
// will use the ID instead of tag filters to identify the instance.
func (e *EC2) Instance() (*ec2.Instance, error) {
	if e.svc == nil {
		e.svc = ec2instances.NewSession(e.Profile, e.Region)
	}

	instance, err := ec2instances.GetInstance(
		e.svc,
		e.ID,
		EC2TagSeparator,
		e.IncludeTags,
		e.ExcludeTags,
	)

	if err != nil {
		return nil, fmt.Errorf("looking for instance in region '%s': %w", e.Region, err)
	}

	if e.ID == "" {
		e.ID = *instance.InstanceId
	}

	return instance, nil
}

// Retrieve returns the administrator credentials for this instance or an error if unable to retrieve them. If the
// getcred field is not set it returns empty strings and no error. This method allows the EC2 host to also implement
// the Cred interface and be called for authentication.
func (e *EC2) Retrieve() (string, string, error) {
	if !e.GetCred {
		return "", "", nil
	}

	instance, err := e.Instance()

	if err != nil {
		return "", "", fmt.Errorf("getting ec2 instance: %s", err)
	}

	keyData, err := ec2instances.ReadPrivateKey(
		viper.GetString("ssh-directory"),
		*instance.KeyName,
	)

	if err != nil {
		return "", "", fmt.Errorf("reading private key: %s", err)
	}

	// Get instance password
	password, err := ec2instances.GetPassword(
		e.svc,
		*instance.InstanceId,
		keyData,
	)
	if err != nil {
		return "", "", fmt.Errorf("getting ec2 administrator credentials: %s", err)
	}

	username := "Administrator"

	return username, password, nil
}
