package config2

import "fmt"

// Settings is the configuration of .RDP file settings.
// https://docs.microsoft.com/en-us/windows-server/remote/remote-desktop-services/clients/rdp-files
type Settings struct {
	Height int `mapstructure:"height"`
	Width  int `mapstructure:"width"`
	Scale  int `mapstructure:"scale"`
}

// Validate returns an error if a config field is invalid.
func (s Settings) Validate() error {
	if s.Width != 0 && (s.Width < 200 || s.Width > 8192) {
		return fmt.Errorf("width value is %d invalid, must be above 200 and below 8192\n", s.Width)
	}

	if s.Height != 0 && (s.Height < 200 || s.Height > 8192) {
		return fmt.Errorf("height value %d is invalid, must be above 200 and below 8192\n", s.Height)
	}

	if s.Scale != 0 && func() bool {
		// Scale is not in list of valid values
		for _, v := range []int{100, 125, 150, 175, 200, 250, 300, 400, 500} {
			if s.Scale == v {
				return false
			}
		}
		return true
	}() {
		return fmt.Errorf("scale value %d is invalid, must be one of 100, 125, 150, 175, 200, 250, 300, 400, 500\n", s.Scale)
	}

	return nil
}
