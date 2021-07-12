package config2

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/danhale-git/runrdp/internal/config2/hosts"

	"github.com/spf13/viper"

	"github.com/danhale-git/runrdp/internal/mock"
)

func vipersFromString(s string) (map[string]*viper.Viper, error) {
	return vipersFromStrings([]string{s})
}

func vipersFromStrings(s []string) (map[string]*viper.Viper, error) {
	readers := readersFromStrings(s)

	return ReadConfigs(readers)
}

func readersFromStrings(s []string) map[string]io.Reader {
	readers := make(map[string]io.Reader)
	for i, v := range s {
		readers[fmt.Sprintf("config%d", i+1)] = strings.NewReader(v)
	}

	return readers
}

func TestReadConfigs(t *testing.T) {
	r := readersFromStrings([]string{
		mock.Config,
	})

	v, err := ReadConfigs(r)
	if err != nil {
		t.Errorf("unexpected error returned: %s", err)
	}

	testKey := "host.awsec2.awsec2test"

	if !v["config1"].IsSet(testKey) {
		t.Errorf("key '%s' should be set but is not", testKey)
	}
}

func TestNew(t *testing.T) {
	v, err := vipersFromString(mock.Config)

	c, err := New(v)
	if err != nil {
		t.Fatalf("unexpected error returned: %s", err)
	}

	if len(c.Hosts) == 0 {
		t.Errorf("configuration object has no hosts after parsing")
	}

	for k := range hosts.Map {
		name := fmt.Sprintf("%stest", k)
		if _, ok := c.Hosts[name]; !ok {
			t.Errorf("host with key '%s' was not loaded into the configuration", name)
		}
	}

	if ec2, ok := c.Hosts["awsec2test"].(*hosts.EC2); ok {
		if ec2.Profile != "TESTVALUE" {
			t.Errorf("unexpected value for ec2test.Profile: expected 'TESTVALUE': got '%s'", ec2.Profile)
		}
	} else {
		t.Errorf("unable to convert hosts.awsec2.ec2test to type hosts.EC2")
	}

	v, err = vipersFromString(`[host.basic.test]
    address = "35.178.168.122"
    cred = "mycred"

[host.awsec2.test]
    tunnel = "mytunnel"
    private = true
    getcred = true
    profile = "default"
    region = "eu-west-2"
    includetags = ["mytag;mytagvalue", "Name;MyInstanceName"]
`)
	_, err = New(v)
	if err == nil {
		t.Errorf("no error returned when config has a duplicate key")
	} else if !errors.Is(err, &DuplicateConfigNameError{}) {
		t.Errorf("unexpecred error returned: expected DuplicateConfigNameError: got %T", errors.Unwrap(err))
	}

	v, err = vipersFromString(`[host.awsec2.test]
	tunnel = "mytunnel"
    private = 1234
    getcred = true
    profile = "default"
    region = "eu-west-2"
    includetags = ["mytag;mytagvalue", "Name;MyInstanceName"]
`)
	_, err = New(v)
	if err == nil {
		t.Errorf("no error returned when config has an incorrect field value type")
	} else if !errors.Is(err, &FieldLoadError{}) {
		t.Errorf("unexpecred error returned: expected FieldLoadError: got %T: %s", errors.Unwrap(err), err)
	}
}
