package mock

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Config is a mock of a config with all entry types
const Config = `[cred.awssm.awssmtest]
    usernameid = "TestInstanceUsername"
    passwordid = "TestInstancePassword"
    region = "eu-west-2"
    profile = "default"

[thycotic.settings]
thycotic-url = "testthycotic-url"
thycotic-domain = "testthycotic-domain"

[cred.thycotic.thycotictest]
	secretid = 1234

[host.awsec2.awsec2test]
    id = "i-12345abc"
	tunnel = "mytunnel"
    private = true
    cred = "awssmtest"
	getcred = true
    profile = "TESTVALUE"
    region = "eu-west-2"
    filterjson = """
    [
      {
        "Name": "tag:Name",
        "Values": ["rdp-target"]
      }
    ]
    """

[host.basic.basictest]
	cred = "global"
	proxy = "global" 
	address = "global" 
	port = "global" 
	username = "global" 
	tunnel = "global" 
	settings = "global"     

[tunnel.tunneltest]
    host = "myiphost"
    localport = "3390"
    key = "C:/Users/me/.ssh/key"
    user = "ubuntu"

[settings.settingstest]
	height = 200
	width = 200
	fullscreen = true
	span = true
	public = true
`

// ConfigKeys returns a slice containing all expected mock config keys
func ConfigKeys() []string {
	return []string{
		"cred.awssm.awssmtest",
		"host.awsec2.awsec2test",
		"host.basic.basictest",
		"tunnel.tunneltest",
		"settings.settingstest",
	}
}

// InstancesWithNames returns a slice of ec2.Instance with the given names and a status of 'running'.
func InstancesWithNames(names ...string) []*ec2.Instance {
	instances := make([]*ec2.Instance, len(names))

	for i, n := range names {
		instances[i] = &ec2.Instance{
			Tags: []*ec2.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(n),
				},
			},
			State: &ec2.InstanceState{
				Code: nil,
				Name: aws.String(ec2.InstanceStateNameRunning),
			},
		}
	}

	return instances
}

// HostCred implements creds.Cred and hosts.Host and defines literal credentials or socket for testing purposes.
type HostCred struct {
	Username, Password string
	Address, Port      string
}

// Retrieve returns the Username and Password fields.
func (h *HostCred) Retrieve() (string, string, error) {
	return h.Username, h.Password, nil
}

// Socket returns the Address and Port fields.
func (h *HostCred) Socket() (string, string, error) {
	return h.Address, h.Port, nil
}

// Validate returns an error if a config field is invalid.
func (h HostCred) Validate() error {
	return nil
}

// Host implements hosts.Host and defines literal socket values for testing purposes.
type Host struct {
	Address, Port string
}

// Socket returns the Address and Port fields.
func (h *Host) Socket() (string, string, error) {
	return h.Address, h.Port, nil
}

// Validate returns an error if a config field is invalid.
func (h Host) Validate() error {
	return nil
}

// Cred implements creds.Cred and defines literal credentials for testing purposes.
type Cred struct {
	Username, Password string
}

// Retrieve returns the Username and Password fields.
func (h *Cred) Retrieve() (string, string, error) {
	return h.Username, h.Password, nil
}

// Validate returns an error if a config field is invalid.
func (h Cred) Validate() error {
	return nil
}
