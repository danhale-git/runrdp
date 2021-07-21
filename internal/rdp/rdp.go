package rdp

// DefaultPort is the standard port for RDP connections.
const DefaultPort = "3389"

type RDP struct {
	Username, Password       string
	Address, Port            string
	Width, Height            int
	Fullscreen, Public, Span bool
}
