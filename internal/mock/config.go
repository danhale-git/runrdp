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
    address = "35.178.168.122"
    cred = "testcred"

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
