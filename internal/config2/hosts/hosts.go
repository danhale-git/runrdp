package hosts

// Map is the source of truth for a complete list of implemented host key names and struct functions.
var Map = map[string]func(int) []interface{}{
	"basic":  BasicStructs,
	"awsec2": EC2Structs,
}

// GlobalFieldNames returns a slice of field name strings corresponding to GlobalHostFields
func GlobalFieldNames() []string {
	return []string{
		"cred",
		"proxy",
		"address",
		"port",
		"username",
		"tunnel",
		"settings",
	}
}

// FieldNameIsGlobal returns true if the given name is in the list of global host field names
func FieldNameIsGlobal(name string) bool {
	for _, n := range GlobalFieldNames() {
		if n == name {
			return true
		}
	}

	return false
}
