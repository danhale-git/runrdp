package config2

import (
	"fmt"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/danhale-git/runrdp/internal/config2/creds"

	"github.com/danhale-git/runrdp/internal/config2/hosts"

	"github.com/spf13/viper"

	"github.com/danhale-git/runrdp/internal/mock"
)

func vipersFromString(s string) map[string]*viper.Viper {
	return vipersFromStrings([]string{s})
}

func vipersFromStrings(s []string) map[string]*viper.Viper {
	readers := readersFromStrings(s)
	v, err := ReadConfigs(readers)
	if err != nil {
		log.Fatalf("unexpected error parsing test config: %s", err)
	}
	return v
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

	for _, testKey := range mock.ConfigKeys() {
		if !v["config1"].IsSet(testKey) {
			t.Errorf("expected a config entry with key '%s' but didn't find one", testKey)
		}
	}

}

func TestNew(t *testing.T) {
	v := vipersFromString(mock.Config)

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

	for k := range creds.Map {
		name := fmt.Sprintf("%stest", k)
		if _, ok := c.creds[name]; !ok {
			t.Errorf("cred with key '%s' was not loaded into the configuration", name)
		}
	}

	if _, ok := c.settings["settingstest"]; !ok {
		t.Errorf("settings with key 'settingstest' was not loaded into the configuration")
	}

	if _, ok := c.tunnels["tunneltest"]; !ok {
		t.Errorf("cred with key 'tunneltest' was not loaded into the configuration")
	}
}

func TestConfiguration_HostsSortedByPattern(t *testing.T) {
	v := vipersFromString(`
[host.basic.abc]
	address = "1.2.3.4"
[host.basic.abc123]
	address = "1.2.3.4"
[host.basic.abc12345]
	address = "1.2.3.4"
[host.basic.zbc]
	address = "1.2.3.4"`)
	c, err := New(v)
	if err != nil {
		t.Errorf("unexpected error creating config: %s", err)
	}

	got := c.HostsSortedByPattern("abc")

	expected := []string{"abc", "abc123", "abc12345"}

	if len(got) != len(expected) {
		t.Fatalf("incorrect number of values returned: expected %d (%s): got %d (%s)",
			len(expected), expected, len(got), got)
	}

	for i, g := range got {
		if expected[i] != g {
			t.Fatalf("unexpected values returned: expected %s: got %s", expected, got)
		}
	}
}

func TestConfiguration_HostKeys(t *testing.T) {
	v := vipersFromString(`
[host.basic.abc]
	address = "1.2.3.4"
[host.basic.abc123]
	address = "1.2.3.4"
[host.basic.abc12345]
	address = "1.2.3.4"`)
	c, err := New(v)
	if err != nil {
		t.Errorf("unexpected error creating config: %s", err)
	}

	got := c.HostKeys()

	expected := []string{"abc", "abc123", "abc12345"}

	if len(got) != len(expected) {
		t.Fatalf("incorrect number of values returned: expected %d (%s): got %d (%s)",
			len(expected), expected, len(got), got)
	}
}

func TestConfiguration_HostExists(t *testing.T) {
	v := vipersFromString(`
[host.basic.abc]
	address = "1.2.3.4"`)
	c, err := New(v)
	if err != nil {
		t.Errorf("unexpected error creating config: %s", err)
	}

	if !c.HostExists("abc") {
		t.Errorf("host 'abc' was expected to exist but did not")
	}

	if c.HostExists("xyz") {
		t.Errorf("host 'xyz' was reported to exist but is not configured")
	}
}

func TestConfiguration_HostCredentials(t *testing.T) {
	c, err := New(map[string]*viper.Viper{})
	if err != nil {
		t.Errorf("unexpected error creating config: %s", err)
	}

	cred := &mock.Cred{
		Username: "creduser",
		Password: "credpassword",
	}

	host := &mock.Host{
		Address: "1.2.3.4",
		Port:    "1234",
	}

	hostCred := &mock.HostCred{
		Username: "hostuser",
		Password: "hostpassword",
		Address:  "1.2.3.4",
		Port:     "1234",
	}

	c.Hosts["testhost"] = hostCred
	c.HostGlobals["testhost"] = map[string]string{
		"cred":     "testcred",
		"username": "globaluser",
	}
	c.AddCred("testcred", cred)

	user, pass, err := c.HostCredentials("testhost")
	if err != nil {
		t.Errorf("unexpected error returned getting host credentials: %s", err)
	}

	if user != "globaluser" {
		t.Errorf("unexpected username '%s': expected 'globaluser'",
			user)
	}

	if pass != "hostpassword" {
		t.Errorf("unexpected password '%s': expected 'credpassword'", pass)
	}

	// Remove global username so cred host username takes priority
	c.HostGlobals["testhost"] = map[string]string{
		"cred": "testcred",
	}

	user, pass, err = c.HostCredentials("testhost")
	if err != nil {
		t.Errorf("unexpected error returned getting host credentials: %s", err)
	}

	if user != "hostuser" {
		t.Errorf("unexpected username '%s': expected 'hostuser'",
			user)
	}

	// Remove host implementation of cred, only the global cred reference to the Cred object remains
	c.Hosts["testhost"] = host

	user, pass, err = c.HostCredentials("testhost")
	if err != nil {
		t.Errorf("unexpected error returned getting host credentials: %s", err)
	}

	if user != "creduser" {
		t.Errorf("unexpected username '%s': expected 'creduser'",
			user)
	}
}

func (c *Configuration) AddCred(name string, cred creds.Cred) {
	c.creds[name] = cred
}
