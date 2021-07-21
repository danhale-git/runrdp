package rdp

import "fmt"

// DefaultPort is the standard port for RDP connections.
const DefaultPort = "3389"

type RDP struct {
	Username, Password       string
	Address, Port            string
	Width, Height            int
	Fullscreen, Public, Span bool
}

func (r RDP) String() string {
	return fmt.Sprintf(`Username: %s
Address: %s
Port: %s
Width %d
Fullscreen: %t
Public: %t
Span: %t`,
		r.Username,
		r.Address,
		r.Port,
		r.Width,
		r.Fullscreen,
		r.Public,
		r.Span,
	)
}
