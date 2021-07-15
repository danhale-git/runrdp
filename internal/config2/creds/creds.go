package creds

// Map is the source of truth for a complete list of implemented host key names and struct functions.
var Map = map[string]func() interface{}{
	"awssm":    SecretsManagerStruct,
	"thycotic": ThycoticStruct,
}
