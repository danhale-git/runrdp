package creds

// Cred can return valid credentials used to authenticate an RDP session.
type Cred interface {
	Retrieve() (string, string, error)
	Validate() error
}

// Map is the source of truth for a complete list of implemented host key names and struct functions.
var Map = map[string]func() interface{}{
	"awssm":    SecretsManagerStruct,
	"thycotic": ThycoticStruct,
}
