# Contributing to inframap-d2

Thanks for your interest in contributing! Whether it's a bug report, a new collector, or a documentation fix, all contributions are welcome.

## Code of Conduct

Be respectful. We're all here to build useful tools. Keep discussions constructive, assume good faith, and focus on the technical merits of contributions.

## Reporting Bugs

Open an issue on [GitHub](https://github.com/ThomasCrouzet/inframap-d2/issues) with:

- **Version**: output of `inframap-d2 version`
- **OS/Arch**: e.g., `linux/amd64`, `darwin/arm64`
- **Steps to reproduce**: minimal config and commands to trigger the bug
- **Expected vs actual behavior**
- **Config file** (anonymized — remove real hostnames, IPs, API tokens)

## Proposing Features

Open an issue with the prefix `[Feature]` in the title.

For **new collectors**, include:
- The infrastructure source you want to support (e.g., Nomad, Terraform state)
- How data is accessed (API, CLI command, file parsing)
- What kind of data it provides (servers, services, networks, etc.)

## Development Setup

### Prerequisites

- **Go 1.25+** — [install guide](https://go.dev/doc/install)
- **golangci-lint** — [install guide](https://golangci-lint.run/welcome/install/)
- **D2** (optional, for rendering) — [install guide](https://d2lang.com/tour/install)
- **make** (optional, for convenience targets)

### Clone and build

```bash
git clone https://github.com/ThomasCrouzet/inframap-d2.git
cd inframap-d2
make build
make test
```

### Verify everything works

```bash
make demo           # Generate demo.d2 from bundled testdata
make lint           # Run linter
```

## Code Style

- **Formatting**: `gofmt` / `goimports` — enforced by the linter
- **Linting**: `golangci-lint run` must pass with no errors
- **Error wrapping**: use `fmt.Errorf("context: %w", err)` to preserve error chains
- **Naming conventions**:
  - Collectors: `XxxCollector` (e.g., `PortainerCollector`, `KubernetesCollector`)
  - Test files: `xxx_test.go` next to the source file
  - Test fixtures: `testdata/xxx/` directory with JSON or YAML files
- **Imports**: standard library first, then external packages, then internal packages

## Architecture

### Data flow

```
cmd/generate.go (runGenerate)
  1. config.Load()           — Viper reads inframap.yml
  2. collector.Collect(cfg)  — runs enabled collectors sequentially:
     ├─ AnsibleCollector     — hosts.yml + group_vars/ → servers, system services
     ├─ ComposeCollector     — compose files + .j2 templates → services, ports, networks
     ├─ TailscaleCollector   — tailscale status --json → IPs, devices, online status
     ├─ SystemdCollector     — systemctl → running services
     ├─ KubernetesCollector  — kubectl → pods, services, ingresses
     ├─ ProxmoxCollector     — Proxmox API → VMs, LXC containers
     └─ PortainerCollector   — Portainer API → containers
     then Merge()            — categorizeServices() + buildTypeGroups()
  3. render.RenderD2()       — generates D2 text output
  4. os.WriteFile()          — writes .d2 file
```

### Key patterns

- **Shared accumulation**: All collectors write into one `*model.Infrastructure` struct. Servers are keyed by hostname — later collectors update fields set by earlier ones (e.g., Tailscale enriches Ansible servers with IPs).
- **Graceful fallback**: ComposeCollector tries the compose-go library first, falls back to raw YAML parsing with Jinja2 stripping.
- **Lazy server creation**: Compose, systemd, and Portainer collectors create servers on-the-fly if they weren't defined by Ansible.
- **Registry pattern**: Collectors self-register via `init()` → `Register()`. No manual wiring needed.
- **Test isolation**: Collectors accept a `TestFile` / `TestData` field to bypass live API/CLI calls in tests.

### Types

**Server types** (`model.ServerType`): `production`, `lab`, `local`, `cluster`, `hypervisor`

**Service types** (`model.ServiceType`): `container`, `database`, `app`, `system`, `vm`, `lxc`, `pod`

## Adding a New Collector

This is the most impactful way to contribute. Follow these steps:

### 1. Create the collector file

Create `internal/collector/myservice.go`:

```go
package collector

import "github.com/ThomasCrouzet/inframap-d2/internal/model"

type MyServiceCollector struct {
    // Config fields populated by Configure()
    URL      string
    APIKey   string
    Server   string
    TestFile string // for testing without live API
}
```

### 2. Register in init()

```go
func init() {
    Register(func() RegisteredCollector { return &MyServiceCollector{} })
}
```

This is all it takes to make your collector discoverable — no imports or wiring elsewhere.

### 3. Implement the RegisteredCollector interface

The interface has 5 methods:

```go
type RegisteredCollector interface {
    Metadata() CollectorMetadata
    Enabled(sources map[string]any) bool
    Configure(section map[string]any) error
    Validate() []ValidationError
    Collect(infra *model.Infrastructure) error
}
```

#### Metadata()

Return a descriptor for your collector:

```go
func (c *MyServiceCollector) Metadata() CollectorMetadata {
    return CollectorMetadata{
        Name:        "myservice",           // internal key (used in config)
        DisplayName: "My Service",          // shown in CLI output
        Description: "Collects X from Y",   // one-line description
        ConfigKey:   "myservice",           // YAML key under sources:
        DetectHint:  "myservice-cli",       // binary/file to detect (for init wizard)
    }
}
```

#### Enabled()

Check if this collector should run based on the raw YAML config:

```go
func (c *MyServiceCollector) Enabled(sources map[string]any) bool {
    section, ok := sources["myservice"].(map[string]any)
    if !ok {
        return false
    }
    if enabled, ok := section["enabled"].(bool); ok {
        return enabled
    }
    // Enabled if the section exists with any config
    return len(section) > 0
}
```

#### Configure()

Populate your struct fields from the raw YAML config section:

```go
func (c *MyServiceCollector) Configure(section map[string]any) error {
    if v, ok := section["url"].(string); ok {
        c.URL = v
    }
    if v, ok := section["api_key"].(string); ok {
        c.APIKey = v
    }
    if v, ok := section["server"].(string); ok {
        c.Server = v
    }
    return nil
}
```

#### Validate()

Return validation errors for problematic config (called by `inframap-d2 validate`):

```go
func (c *MyServiceCollector) Validate() []ValidationError {
    var errs []ValidationError
    if c.URL == "" {
        errs = append(errs, ValidationError{
            Field:      "sources.myservice.url",
            Message:    "URL is required",
            Suggestion: "Add url: https://myservice.local to your config",
        })
    }
    return errs
}
```

#### Collect()

The main logic — fetch data and write it into the shared `*model.Infrastructure`:

```go
func (c *MyServiceCollector) Collect(infra *model.Infrastructure) error {
    // Fetch data (or read from TestFile for tests)
    data, err := c.fetchData()
    if err != nil {
        return fmt.Errorf("fetch myservice data: %w", err)
    }

    // Ensure the server exists
    server := infra.EnsureServer(c.Server)
    server.Type = model.ServerTypeLab

    // Add services
    for _, item := range data {
        svc := &model.Service{
            Name:  item.Name,
            Image: item.Image,
            Type:  model.ServiceTypeContainer,
        }
        server.Services = append(server.Services, svc)
    }

    return nil
}
```

### 4. Add test fixtures

Create `testdata/myservice/` with sample JSON or YAML data:

```
testdata/myservice/
└── data.json        # Sample API response
```

### 5. Write tests

Create `internal/collector/myservice_test.go`:

```go
package collector

import (
    "testing"

    "github.com/ThomasCrouzet/inframap-d2/internal/model"
    "github.com/stretchr/testify/assert"
)

func TestMyServiceCollect(t *testing.T) {
    c := &MyServiceCollector{
        URL:      "https://myservice.local",
        Server:   "testhost",
        TestFile: "../../testdata/myservice/data.json",
    }

    infra := model.NewInfrastructure()
    err := c.Collect(infra)

    assert.NoError(t, err)
    assert.Contains(t, infra.Servers, "testhost")
    assert.NotEmpty(t, infra.Servers["testhost"].Services)
}
```

### 6. Optional enhancements

- **Icons**: add entries to `internal/render/icons.go` for services your collector discovers
- **Categories**: add patterns to `internal/render/categories.go` for auto-categorization

### Reference implementations

- **Simple API collector**: `internal/collector/portainer.go` — straightforward HTTP API, single server
- **Complex collector**: `internal/collector/kubernetes.go` — CLI execution, multiple servers, deduplication
- **File-based collector**: `internal/collector/ansible.go` — YAML file parsing, multi-source correlation
- **Registry interface**: `internal/collector/registry.go` — the `RegisteredCollector` interface definition

## Tests

Tests use [testify/assert](https://github.com/stretchr/testify) for assertions.

```bash
# Run all tests
make test

# Run tests for a specific package
go test ./internal/collector/

# Run a specific test function
go test ./internal/collector/ -run TestPortainerCollect

# Verbose output
go test -v ./internal/render/ -run TestD2Renderer
```

- Test files live next to their source: `portainer.go` → `portainer_test.go`
- Test fixtures go in `testdata/xxx/` with JSON or YAML sample data
- Collectors use a `TestFile` field to read fixtures instead of making live API calls
- Run `make lint` before submitting — the linter catches common issues

## Commit Conventions

Use the format `type: description`:

| Type | Use for |
|------|---------|
| `feat` | New feature or collector |
| `fix` | Bug fix |
| `refactor` | Code restructuring without behavior change |
| `test` | Adding or improving tests |
| `docs` | Documentation changes |
| `chore` | Build, CI, dependency updates |

Examples:

```
feat: add Nomad collector
fix: handle empty Tailscale peer list
refactor: extract common HTTP client for API collectors
test: add fixtures for Proxmox LXC containers
docs: document systemd collector SSH setup
```

## Pull Request Process

1. **Fork** the repository on [GitHub](https://github.com/ThomasCrouzet/inframap-d2)
2. **Create a feature branch** from `main`: `git checkout -b feat/nomad-collector`
3. **Make your changes** with tests and passing lint
4. **Push** to your fork: `git push origin feat/nomad-collector`
5. **Open a pull request** against `main` with:
   - A clear description of what the PR does
   - Any relevant issue numbers
   - How to test the changes

Keep PRs focused — one feature or fix per PR. Large changes are easier to review when split into smaller PRs.

## Release Process

For reference — releases are managed by maintainers:

- **Semantic versioning**: `vMAJOR.MINOR.PATCH`
- **GoReleaser**: triggered by git tags (`git tag v0.4.0 && git push --tags`)
- **Platforms**: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`
- **Artifacts**: tar.gz archives + checksums published to GitHub releases
