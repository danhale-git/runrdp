package hosts

// Map is the source of truth for a complete list of implemented host key names and struct functions.
var Map = map[string]func(int) []interface{}{
	"basic":  BasicStructs,
	"awsec2": EC2Structs,
}
