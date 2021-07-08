package config2

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/danhale-git/runrdp/internal/mock"
)

func loadConfig(raw string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("toml")
	if err := v.ReadConfig(strings.NewReader(raw)); err != nil {
		panic(fmt.Sprintf("TEST CONFIG LOAD FAILED: %s, %T", err, errors.Unwrap(err)))
	}

	return v, nil
}

/*func TestRead(t *testing.T) {
	v, err := Load(bytes.NewBuffer([]byte(mock.Config)))
	if err != nil {
		t.Errorf("unexpected error returned: %s", err)
	}

	if v == nil {
		t.Errorf("returned viper.Viper instance is nil but no error was returned")
	}

	_, err = Load(bytes.NewBuffer([]byte(mock.ConfigWithDuplicate)))
	if err == nil {
		t.Errorf("no error returned when config has a duplicate key")
	} else if !errors.Is(errors.Unwrap(err), viper.ConfigParseError{}) {
		t.Errorf("unexpected error returned when config has a duplicate key: expected viper.ConfigParseError: got %T", errors.Unwrap(err))
	}

}*/

func TestParse(t *testing.T) {
	v, err := loadConfig(mock.Config)
	if err != nil {
		t.Errorf("error loading config: %s", err)
	}

	c, err := Parse(v)
	if err != nil {
		t.Errorf("unexpected error returned: %s", err)
	}

	if len(c.Hosts) == 0 {
		t.Errorf("configuration object has no hosts after parsing")
	}

	v, err = loadConfig(mock.ConfigWithDuplicate)
	_, err = Parse(v)
	if err == nil {
		t.Errorf("no error returned when config has a duplicate key")
	} else if !errors.Is(err, &DuplicateConfigNameError{}) {
		t.Errorf("unexpecred error returned: expected DuplicateConfigNameError: got %T", errors.Unwrap(err))
	}

	// TODO: need to make sure config is validated
	/*v, err = loadConfig(mock.ConfigWithUnknownField)
	_, err = Parse(v)
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
