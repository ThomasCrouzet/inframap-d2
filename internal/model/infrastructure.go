package model

// Infrastructure is the top-level aggregate of all discovered infra.
type Infrastructure struct {
	Servers      map[string]*Server
	ServerGroups map[string]*ServerGroup
	Devices      map[string]*Device
	Networks     map[string]*Network
	TailnetName  string
}

// NewInfrastructure creates an initialized Infrastructure.
func NewInfrastructure() *Infrastructure {
	return &Infrastructure{
		Servers:      make(map[string]*Server),
		ServerGroups: make(map[string]*ServerGroup),
		Devices:      make(map[string]*Device),
		Networks:     make(map[string]*Network),
	}
}

// ServerGroup groups servers by type or role.
type ServerGroup struct {
	Name    string
	Label   string
	Servers []string // hostnames
}
