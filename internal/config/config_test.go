package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func newConfiguration(configurations ...*viper.Viper) Configuration {
	c := Configuration{}
	c.Data = make(map[string]*viper.Viper)
	c.Hosts = make(map[string]Host)
	c.Creds = make(map[string]*Cred)
	c.creds = make(map[string]Cred)

	for i, config := range configurations {
		c.Data["config_"+string(i)] = config
	}

	return c
}

func configFromString(fileContents string) (*viper.Viper, error) {
	r := strings.NewReader(fileContents)
	v := viper.New()
	v.SetConfigType("toml")
	err := v.ReadConfig(r)

	return v, err
}

func TestConfiguration_BuildData(t *testing.T) {
	testConfigString := strings.Join(
		[]string{
			hostTestData(),
			credTestData(),
		},
		"\n",
	)

	testConfig, err := configFromString(testConfigString)
	if err != nil {
		t.Errorf("error reading config data: %s", err)
		return
	}

	c := newConfiguration(
		testConfig,
	)

	c.BuildData()

	testHosts(t, c.Hosts)
	testCreds(t, c.creds)
}
