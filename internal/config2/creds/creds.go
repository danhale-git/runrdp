package creds

// Cred can return valid credentials used to authenticate an RDP session.
type Cred interface {
	Retrieve() (string, string, error)
}
