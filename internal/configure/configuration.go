package configure

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Configuration struct {
	Files    map[string]*viper.Viper
	Desktops map[string]Desktop
}

func (c *Configuration) Get(key string) (interface{}, error) {
	for _, config := range c.Files {
		if config.IsSet(key) {
			return config.Get(key), nil
		}
	}

	return nil, fmt.Errorf("config entry '%s' not found", key)
}

func (c *Configuration) ReadLocalConfigFiles() {
	c.Files = make(map[string]*viper.Viper)
	c.Desktops = make(map[string]Desktop)

	c.loadConfigFiles()
	c.loadDesktopConfigs()
}

func (c *Configuration) loadConfigFiles() {
	for _, configFile := range configFileNames() {
		c.Files[configFile] = loadFile(configFile)
	}
}

func (c *Configuration) loadDesktopConfigs() {
	c.Desktops = LoadDesktopConfigurations(
		c.getConfig("cred"),
		c.getConfig("host"),
	)
}

func configFileNames() []string {
	files, err := ioutil.ReadDir(viper.GetString("config-root"))
	if err != nil {
		log.Fatal(err)
	}

	names := make([]string, 0)

	for _, f := range files {
		n := f.Name()
		if strings.TrimSpace(n) == "" {
			continue
		}

		names = append(names, n)
	}

	return names
}

func loadFile(name string) *viper.Viper {
	newViper := viper.New()

	newViper.SetConfigType("toml")
	newViper.SetConfigFile(filepath.Join(
		viper.GetString("config-root"),
		name,
	))

	if err := newViper.ReadInConfig(); err == nil {
		fmt.Printf("%s config loaded from %s\n", name, newViper.ConfigFileUsed())
	} else {
		fmt.Printf("failed to load %s config: %v\n", name, err)
	}

	return newViper
}

func (c *Configuration) getConfig(key string) map[string]map[string]interface{} {
	var allConfigs = make(map[string]map[string]interface{})

	for _, config := range c.Files {
		for kind, items := range config.GetStringMap(key) {
			if _, ok := allConfigs[kind]; !ok {
				allConfigs[kind] = make(map[string]interface{})
			}

			for k, v := range items.(map[string]interface{}) {
				allConfigs[kind][k] = v
			}
		}
	}

	return allConfigs
}
