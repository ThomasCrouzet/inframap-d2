package collector

import "github.com/ThomasCrouzet/inframap-d2/internal/model"

// RegisteredCollector defines the interface for self-registering data sources.
type RegisteredCollector interface {
	Metadata() CollectorMetadata
	Enabled(sources map[string]any) bool
	Configure(section map[string]any) error
	Validate() []ValidationError
	Collect(infra *model.Infrastructure) error
}

// CollectorMetadata describes a collector for discovery and documentation.
type CollectorMetadata struct {
	Name        string // internal key, e.g. "ansible"
	DisplayName string // human-readable, e.g. "Ansible Inventory"
	Description string // one-line description
	ConfigKey   string // YAML key under sources, e.g. "ansible"
	DetectHint  string // filesystem hint for auto-detection, e.g. "hosts.yml"
}

// ValidationError reports a config problem with a suggested fix.
type ValidationError struct {
	Field      string // dotted path, e.g. "sources.ansible.inventory"
	Message    string // what's wrong
	Suggestion string // how to fix it
}

var registry []func() RegisteredCollector

// Register adds a collector factory to the global registry.
// Each collector calls this in its init().
func Register(factory func() RegisteredCollector) {
	registry = append(registry, factory)
}

// All returns fresh instances of every registered collector.
func All() []RegisteredCollector {
	out := make([]RegisteredCollector, len(registry))
	for i, f := range registry {
		out[i] = f()
	}
	return out
}
