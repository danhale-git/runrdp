package hosts

// BasicStruct a struct of type hosts.Basic.
func BasicStruct() interface{} {
	return &Basic{}
}

// TODO: Implement this
func (b *Basic) ValidateBasic() {

}

// Basic defines a host to connect to using an IP or hostname.
//
// Users may configure a basic host with global fields only (see config.toml). No other fields are defined.
type Basic struct{}

// Socket returns this host's IP or hostname.
func (h *Basic) Socket() (string, string, error) {
	return "", "", nil
}
