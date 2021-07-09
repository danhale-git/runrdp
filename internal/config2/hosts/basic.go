package hosts

// BasicStructs returns a slice containing structs of type hosts.Basic with the given length.
func BasicStructs(l int) []interface{} {
	structs := make([]interface{}, l)
	for i := range structs {
		structs[i] = &Basic{}
	}

	return structs
}

// TODO: Implement this
func ValidateBasic() {

}

// Basic defines a host to connect to using an IP or hostname.
//
// Users may configure a basic host with global fields only (see config.toml). No other fields are defined.
type Basic struct{}

// Socket returns this host's IP or hostname.
func (h *Basic) Socket() (string, string, error) {
	return "", "", nil
}
