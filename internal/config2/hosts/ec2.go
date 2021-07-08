package hosts

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/danhale-git/runrdp/internal/aws/ec2instances"
	"github.com/spf13/viper"
)

func ParseEC2(v *viper.Viper) (map[string]interface{}, []interface{}, error) {
	key := "host.awsec2"
	if !v.IsSet(key) {
		return nil, nil, nil
	}
	raw := v.Get(key).(map[string]interface{})
	structs := make([]interface{}, len(raw))
	for i := range structs {
		structs[i] = &EC2{}
	}

	/*for _, v := range raw {
			// TODO: call validation function
	}*/

	return raw, structs, nil
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
func (h *EC2) Socket() (string, string, error) {
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
func (h *EC2) Instance() (*ec2.Instance, error) {
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
		return nil, err
	}

	if h.ID == "" {
		h.ID = *instance.InstanceId
	}

	return instance, nil
}

// Retrieve returns the administrator credentials for this instance or an error if unable to retrieve them. If the
// getcred field is not set it returns empty strings and no error. This method allows the EC2 host to also implement
// the Cred interface and be called for authentication.
func (h *EC2) Retrieve() (string, string, error) {
	if !h.GetCred {
		return "", "", nil
	}

	instance, err := h.Instance()

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
		h.svc,
		*instance.InstanceId,
		keyData,
	)
	if err != nil {
		return "", "", fmt.Errorf("getting ec2 administrator credentials: %s", err)
	}

	username := "Administrator"

	return username, password, nil
}
