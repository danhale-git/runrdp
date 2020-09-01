package desktops

import (
	"encoding/json"
	"fmt"
)

type HostType int

const (
	IP HostType = iota
	EC2
)

type CredentialType int

const (
	Manual HostType = iota
	EC2WindowsPassword
	AWSSecretsManager
)

type Config struct {
	Desktops []DesktopConfig
}

type DesktopConfig struct {
	Host HostConfig
}

type HostConfig struct {
	Type string
	Name string
	ID   string
	Port int
}

type CredsConfig struct {
	Type string
	ID   string
}

func LoadDesktops(config Config) {
	s, _ := json.MarshalIndent(config, "", "  ")
	fmt.Println(string(s))
}
