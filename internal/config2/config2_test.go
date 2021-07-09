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
		t.Errorf("unexpected error returned: %s", err)
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

	v, err = vipersFromString(mock.ConfigWithDuplicate)
	_, err = New(v)
	if err == nil {
		t.Errorf("no error returned when config has a duplicate key")
	} else if !errors.Is(err, &DuplicateConfigNameError{}) {
		t.Errorf("unexpecred error returned: expected DuplicateConfigNameError: got %T", errors.Unwrap(err))
	}

	// TODO: need to make sure config is validated
	/*v, err = loadConfig(mock.ConfigWithUnknownField)
	_, err = New(v)
	if err == nil {
		t.Errorf("no error returned when config has an unknown field")
	} else if !errors.Is(err, &DuplicateConfigNameError{}) {
		t.Errorf("unexpecred error returned: expected DuplicateConfigNameError: got %T", errors.Unwrap(err))
	}*/

	/*if c == nil {
		return
	}
	fmt.Println(len(c.Hosts))
	for _, h := range c.Hosts {
		fmt.Printf("%+v\n", h)
	}*/
}
