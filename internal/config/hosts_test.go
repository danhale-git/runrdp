package config

import (
	"testing"
)

// hostTestData defines the raw data which constructs the Host structs in the hosts parameter of testHosts.
func hostTestData() string {
	return `
[host.basic.basichosttest]
	address = "35.178.168.122"
	cred = "mycred"

[host.awsec2.ec2hosttest]
    id = "i-07dd0954800829f3b"
    private = false
    getcred = true
    profile = "default"
    region = "eu-west-2"
    includetags = ["Name:MyInstanceName"]
	cred = "mycred"

[cred.awssm.mycred]
    usernameid = "TestInstanceUsername"
    passwordid = "TestInstancePassword"
    region = "eu-west-2"`
}

// testHosts is called in config_test.go
func testHosts(t *testing.T, hosts map[string]Host) {
	testBasicHost(t, hosts)
	testEC2Host(t, hosts)
}

func testBasicHost(t *testing.T, hosts map[string]Host) {
	h, ok := hosts["basichosttest"]
	if !ok {
		t.Errorf("BasicHost is not in Configuration.Hosts")
		return
	}

	_, ok = h.(interface{}).(*BasicHost)
	if !ok {
		t.Errorf("BasicHost cannot be converted to type BasicHost:\n%+v", h)
	}
}

func testEC2Host(t *testing.T, hosts map[string]Host) {
	h, ok := hosts["ec2hosttest"]
	if !ok {
		t.Errorf("EC2Host is not in Configuration.Hosts")
		return
	}

	_, ok = h.(interface{}).(*EC2Host)

	if !ok {
		t.Errorf("EC2Host cannot be converted to type EC2Host:\n%+v", h)
	}
}
