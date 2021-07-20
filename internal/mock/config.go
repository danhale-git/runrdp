package mock

// Config is a mock of a config with all entry types
const Config = `[cred.awssm.awssmtest]
    usernameid = "TestInstanceUsername"
    passwordid = "TestInstancePassword"
    region = "eu-west-2"
    profile = "default"

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
    includetags = ["mytag;mytagvalue", "Name;MyInstanceName"]
    excludetags = ["mytag;myothervalue"]

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
	scale = 200
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

func (h Cred) Validate() error {
	return nil
}
