package config

// Tunnel has the details for opening an 'SSH tunnel' (SSH port forwarding) including a reference to a Host config which
// will be the forwarding server.
type Tunnel struct {
	Host      string `mapstructure:"host"`
	LocalPort string `mapstructure:"localport"`
	Key       string `mapstructure:"key"`
	User      string `mapstructure:"user"`
}

// Validate returns an error if a config field is invalid.
func (t Tunnel) Validate() error {
	return nil
}
