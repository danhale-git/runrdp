package hosts

// Map is the source of truth for a complete list of implemented host key names and struct functions.
var Map = map[string]func() interface{}{
	"basic":  BasicStruct,
	"awsec2": EC2Struct,
}

// Global fields may be used with any host type.
const (
	GlobalCred GlobalFields = iota
	GlobalProxy
	GlobalAddress
	GlobalPort
	GlobalUsername
	GlobalTunnel
	GlobalSettings
)

// GlobalFields are the names of fields which may be configured in any host.
type GlobalFields int

func (p GlobalFields) String() string {
	return GlobalFieldNames()[p]
}

// GlobalFieldNames returns a slice of field name strings corresponding to GlobalFields
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
