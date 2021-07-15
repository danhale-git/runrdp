package config2

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/danhale-git/runrdp/internal/config2/creds"

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

	for _, testKey := range mock.ConfigKeys() {
		if !v["config1"].IsSet(testKey) {
			t.Errorf("expected a config entry with key '%s' but didn't find one", testKey)
		}
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
