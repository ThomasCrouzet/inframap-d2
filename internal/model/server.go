package model

// ServerType classifies a server's role.
type ServerType string

const (
	ServerTypeProduction ServerType = "production"
	ServerTypeLab        ServerType = "lab"
	ServerTypeLocal      ServerType = "local"
	ServerTypeCluster    ServerType = "cluster"
	ServerTypeHypervisor ServerType = "hypervisor"
)

// Server represents a physical or virtual machine.
type Server struct {
	Hostname      string
	Label         string
	PublicIP      string
	TailscaleIP   string
	Type          ServerType
	OS            string
	Online        bool
	AnsibleGroups []string
	Services      []*Service
}

// AddService appends a service to this server.
func (s *Server) AddService(svc *Service) {
	s.Services = append(s.Services, svc)
}
