package config

import (
	"testing"
)

// credTestData defines the raw data which constructs the Cred structs in the creds parameter of testCreds.
func credTestData() string {
	return `
[cred.awssm.secretsmanagertest]
    usernameid = "TestInstanceUsername"
    passwordid = "TestInstancePassword"
    region = "eu-west-2"`
}

// testCreds is called in config_test.go.
func testCreds(t *testing.T, creds map[string]Cred) {
	testSecretsmanagerCred(t, creds)
}

func testSecretsmanagerCred(t *testing.T, creds map[string]Cred) {
	c, ok := creds["secretsmanagertest"]
	if !ok {
		t.Errorf("SecretsManagerCred is not in Configuration.Creds")
	}

	_, ok = c.(interface{}).(*SecretsManagerCred)
	if !ok {
		t.Errorf("SecretsManagerCred Cred cannot be converted to type SecretsManagerCred")
	}
}
