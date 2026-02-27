package model

// Network represents a Docker network.
type Network struct {
	Name     string
	Driver   string
	Services []string // service names connected to this network
}

// Connection represents a link between two entities.
type Connection struct {
	From  string
	To    string
	Label string
	Style string
}
