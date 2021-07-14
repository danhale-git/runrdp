package config2

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"reflect"
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

	if awsec2test, ok := c.Hosts["awsec2test"].(*hosts.EC2); ok {
		checkFields(t, awsec2test)
	} else {
		t.Errorf("failed to get or convert type *hosts.EC2")
	}

	if awssmtest, ok := c.creds["awssmtest"].(*creds.SecretsManager); ok {
		checkFields(t, awssmtest)
	} else {
		t.Errorf("failed to get or convert type *creds.SecretsManager")
	}

	if thycotictest, ok := c.creds["thycotictest"].(*creds.Thycotic); ok {
		checkFields(t, thycotictest)
	} else {
		t.Errorf("unable to convert awsec2test to type *creds.Thycotic")
	}

	settingstest := c.settings["settingstest"]
	checkFields(t, &settingstest)

	tunneltest := c.tunnels["tunneltest"]
	checkFields(t, &tunneltest)

	// Basic doesn't have any fields so we use it to test global fields
	for _, g := range hosts.GlobalFieldNames() {
		if globalVal, ok := c.HostGlobals["basictest"][g]; ok {
			if globalVal != "global" {
				t.Errorf("config basictest has unexpected value for global field %s: expected 'global': got '%s'",
					g, globalVal)
			}
		} else {
			t.Errorf("config for basictest is missing global field %s", g)
		}
	}

	v, err = vipersFromString(`
[host.basic.test]
    address = "35.178.168.122"
    cred = "mycred"

[host.awsec2.test]
    tunnel = "mytunnel"
    private = true
    getcred = true
    profile = "default"
    region = "eu-west-2"
    includetags = ["mytag;mytagvalue", "Name;MyInstanceName"]`)
	_, err = New(v)
	if err == nil {
		t.Errorf("no error returned when config has a duplicate key")
	} else if !errors.Is(err, &DuplicateConfigNameError{}) {
		t.Errorf("unexpecred error returned: expected DuplicateConfigNameError: got %T", errors.Unwrap(err))
	}

	v, err = vipersFromString(`
[host.awsec2.test]
	tunnel = "mytunnel"
    private = 1234
    getcred = true
    profile = "default"
    region = "eu-west-2"
    includetags = ["mytag;mytagvalue", "Name;MyInstanceName"]`)
	_, err = New(v)
	if err == nil {
		t.Errorf("no error returned when config has an incorrect field value type")
	} else if !errors.Is(err, &FieldLoadError{}) {
		t.Errorf("unexpecred error returned: expected FieldLoadError: got %T: %s", errors.Unwrap(err), err)
	}
}

func checkFields(t *testing.T, str interface{}) {
	value := reflect.ValueOf(str).Elem()
	if zero, name := structHasZeroField(value); zero {
		t.Errorf("%s field %s has a zero value.", value.Type().Name(), name)
	}
}

func structHasZeroField(values reflect.Value) (bool, string) {
	structType := values.Type()

	for i := 0; i < structType.NumField(); i++ {
		// It is not possible to check unexported fields
		if structType.Field(i).PkgPath != "" {
			continue
		}

		v := values.Field(i)

		if isZero(v) {
			return true, structType.Field(i).Name
		}
	}

	return false, ""
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanSet() {
				z = z && isZero(v.Field(i))
			}
		}
		return z
	case reflect.Ptr:
		return isZero(reflect.Indirect(v))
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	result := v.Interface() == z.Interface()

	return result
}
