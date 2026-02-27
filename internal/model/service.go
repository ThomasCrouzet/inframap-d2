package model

// ServiceType classifies a service.
type ServiceType string

const (
	ServiceTypeContainer ServiceType = "container"
	ServiceTypeDatabase  ServiceType = "database"
	ServiceTypeApp       ServiceType = "app"
	ServiceTypeSystem    ServiceType = "system"
	ServiceTypeVM        ServiceType = "vm"
	ServiceTypeLXC       ServiceType = "lxc"
	ServiceTypePod       ServiceType = "pod"
)

// Service represents a container, application, or system service.
type Service struct {
	Name        string
	Image       string
	Type        ServiceType
	Ports       []PortMapping
	Networks    []string
	DependsOn   []string
	Volumes     []VolumeMount
	HealthCheck *HealthCheck
	ComposeFile string
	Category    string // for grouping (media, productivity, infra, etc.)
}

// VolumeMount represents a volume binding.
type VolumeMount struct {
	Source string
	Target string
}

// HealthCheck represents a service health check.
type HealthCheck struct {
	Port           int
	Path           string
	ExpectedStatus int
	Timeout        int
}
