# inframap-d2

A CLI tool that auto-generates [D2](https://d2lang.com) infrastructure diagrams from your existing config files — Ansible inventories, Docker Compose files, Tailscale networks, Kubernetes clusters, Proxmox VE, Portainer, and systemd services.

**One command to map your homelab, self-hosted stack, or production infrastructure.**

## Get Started

### 1. Install

**Pre-built binaries** (linux, macOS — amd64/arm64):

Download from the [Releases](https://github.com/ThomasCrouzet/inframap-d2/releases) page.

**Or build from source** (requires Go 1.25+):

```bash
go install github.com/ThomasCrouzet/inframap-d2@latest
```

**Optional**: install [D2](https://d2lang.com/tour/install) to render diagrams to SVG/PNG.

### 2. Create your config

The fastest way — the interactive wizard scans your machine for Ansible inventories, Docker Compose files, and Tailscale, then generates a config:

```bash
inframap-d2 init
```

Or copy the example and edit manually:

```bash
cp inframap.example.yml inframap.yml
# Edit paths to match your infrastructure
```

### 3. Generate your diagram

```bash
inframap-d2 generate -o infra.d2 --render
open infra.svg
```

That's it. You now have an SVG diagram of your entire infrastructure.

## Try the Demo

No setup needed — the repository includes sample data so you can see the tool in action immediately:

```bash
git clone https://github.com/ThomasCrouzet/inframap-d2.git
cd inframap-d2
make demo-svg
open demo.svg
```

## What It Produces

inframap-d2 generates a `.d2` file describing your infrastructure as a nested diagram:

```
direction: right

tailnet: "Tailscale — my-tailnet" {
  production: "Production" {                    # Servers grouped by type
    gateway: "gateway — 203.0.113.10" {         # Public IP shown
      web: "web :3000" { icon: ... }            # Services with ports and icons
      db: "PostgreSQL" { shape: cylinder }       # Databases as cylinders
    }
  }
  lab: "Lab Servers" {
    atlas: "atlas" {
      uptime-kuma: "uptime-kuma :3001" { ... }
      it-tools: "it-tools :7100" { ... }
    }
  }
  local: "Local" {
    minicore: "minicore" {
      media: "Media" {                          # Services grouped by category
        radarr: "radarr :7878" { ... }
        sonarr: "sonarr :8989" { ... }
      }
    }
  }
  devices: "Other Devices" {                    # Tailscale peers (phones, laptops)
    user-phone: "user-phone" { icon: ... }
  }
}

cloudflare -> tailnet.production.gateway        # External connections
web -> db { style.stroke-dash: 3 }              # depends_on relationships
```

Render to SVG with `d2 infra.d2 infra.svg` or use `--render` to auto-render.

## Supported Sources

| Source | What it collects | Input |
|--------|-----------------|-------|
| **Ansible** | Servers, groups, system services | `hosts.yml` + `group_vars/` |
| **Docker Compose** | Containers, ports, networks, dependencies | `docker-compose.yml` (+ Jinja2 `.j2` templates) |
| **Tailscale** | VPN peers, IPs, online status, devices | `tailscale status --json` or JSON file |
| **systemd** | Running services | `systemctl` (local or via SSH) |
| **Kubernetes** | Pods, services, ingresses | `kubectl` with kubeconfig |
| **Proxmox VE** | VMs, LXC containers | REST API with token |
| **Portainer** | Docker containers | REST API with key |

You only need to configure the sources you use. All sources are optional.

## Configuration

Full YAML reference with all options:

```yaml
output: infrastructure.d2
layout: dagre            # D2 layout engine
direction: right         # "right" (horizontal) or "down" (vertical)
theme: default           # default, dark, monochrome, ocean

sources:
  # Ansible inventory — servers, groups, system services
  ansible:
    inventory: ./inventory/hosts.yml
    group_vars: ./inventory/group_vars
    primary_group: tailnet       # Group to use as primary server source

  # Docker Compose — containers, ports, networks, dependencies
  compose:
    files:
      - path: ./docker-compose.yml
        server: myserver         # Hostname to assign services to
      - path: ./templates/compose.yml.j2
        server: myserver
        template: true           # Strip Jinja2 {{ vars }} before parsing
    scan_dirs:
      - path: ~/docker
        server: homelab          # Recursively find compose files in this directory

  # Tailscale — VPN peers, IPs, online status, devices
  tailscale:
    enabled: true
    json_file: ""                # Optional: path to `tailscale status --json` output
    include_offline: false       # Include offline peers in the diagram

  # systemd — running services from local or remote servers
  systemd:
    servers:
      - host: myserver           # Hostname in the diagram
        ssh: admin@192.168.1.10  # SSH target (omit for local)
        filter: [nginx, postgres, redis]  # Only include these (substring match)
        exclude: [snapd, fwupd]  # Exclude these (substring match)

  # Kubernetes — pods, services, ingresses
  kubernetes:
    kubeconfig: ~/.kube/config
    context: my-cluster          # K8s context name (optional)
    namespaces: []               # Filter namespaces (empty = all)

  # Proxmox VE — VMs and LXC containers
  proxmox:
    api_url: https://pve.local:8006
    token_id: user@pam!inframap  # API token ID
    token: xxxx-xxxx-xxxx        # API token secret
    insecure: false              # Skip TLS verification

  # Portainer — Docker containers via API
  portainer:
    url: https://portainer.local:9443
    api_key: ptr_xxxx            # API key from User Settings
    endpoint: 1                  # Portainer endpoint ID
    server: docker-host          # Hostname to assign containers to

display:
  show_devices: true             # Show non-server Tailscale peers (phones, laptops)
  show_volumes: false            # Show volume mounts
  group_by: category             # Group services by category

render:
  detail_level: standard         # minimal, standard, detailed
  auto_render: false             # Auto-render to SVG/PNG after generating
  format: svg                    # svg or png
```

Secrets can also be set via environment variables:
- `INFRAMAP_PORTAINER_API_KEY`
- `INFRAMAP_PROXMOX_TOKEN_ID`
- `INFRAMAP_PROXMOX_TOKEN`

See [`inframap.example.yml`](inframap.example.yml) for a real-world example.

## Commands

### `generate`

Collect infrastructure data and generate a D2 diagram.

```bash
# Using config file (most common)
inframap-d2 generate -o infra.d2

# Generate and render to SVG in one step
inframap-d2 generate -o infra.d2 --render

# Override theme and detail level
inframap-d2 generate -o infra.d2 --theme dark --detail detailed

# Flags only, no config file
inframap-d2 generate \
  --ansible-inventory ./inventory/hosts.yml \
  --compose-scan-dir ~/docker:myserver \
  --tailscale \
  -o infra.d2
```

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | `-c` | Config file path (default: `inframap.yml`) |
| `--output` | `-o` | Output D2 file (default: `infrastructure.d2`) |
| `--detail` | | `minimal`, `standard`, `detailed` |
| `--theme` | | `default`, `dark`, `monochrome`, `ocean` |
| `--render` | | Auto-render to SVG/PNG (requires `d2`) |
| `--format` | | `svg` or `png` (default: `svg`) |
| `--ansible-inventory` | | Path to Ansible `hosts.yml` |
| `--ansible-group-vars` | | Path to Ansible `group_vars/` |
| `--compose-file` | | Compose file (format: `path:server`, repeatable) |
| `--compose-scan-dir` | | Directory to scan (format: `path:server`, repeatable) |
| `--tailscale` | | Enable Tailscale collection |
| `--tailscale-json` | | Path to Tailscale status JSON file |

### `init`

Create an `inframap.yml` interactively. The wizard detects your environment and walks you through each source.

```bash
inframap-d2 init
```

### `validate`

Check that all configured sources are valid — files exist, binaries are available, APIs reachable.

```bash
inframap-d2 validate
inframap-d2 validate -c custom-config.yml
```

### `version`

```bash
inframap-d2 version
```

## Themes & Detail Levels

### Themes

| Theme | Description |
|-------|-------------|
| `default` | Colorful pastels — production (red), lab (green), local (yellow) |
| `dark` | Dark backgrounds with bright accents |
| `monochrome` | Grayscale, print-friendly |
| `ocean` | Blue tones with red production highlights |

### Detail Levels

| Level | What's shown |
|-------|-------------|
| `minimal` | Servers and groups only |
| `standard` | Services, ports, icons, devices, external connections (default) |
| `detailed` | Everything expanded: all system services, full metadata |

## Source Details

### Ansible

- Servers are read from the `primary_group` hosts
- `server_type` host variable sets the type (`production`, `lab`, `local`)
- System services (netdata, cockpit) discovered from `group_vars/all.yml`
- Health checks from `group_vars/{group}/vars.yml`
- `bootstrap` group provides public IP mapping

### Docker Compose

- Uses [compose-go](https://github.com/compose-spec/compose-go) with a fallback to raw YAML parsing
- Jinja2 templates (`.j2`): `{{ variables }}` are replaced with placeholders before parsing
- Database images (postgres, mysql, redis, mongo, etc.) get `shape: cylinder`
- `scan_dirs` recursively finds `docker-compose.yml`, `compose.yml` and their `.yaml` variants
- `depends_on` relationships render as dashed connections

### Tailscale

- Runs `tailscale status --json` live, or reads from a JSON file
- Peers tagged `tag:server` create or update server entries
- Other peers (phones, laptops, IoT) appear in the "Other Devices" section
- Enriches servers from other sources with Tailscale IPs and online status

### systemd

- Runs `systemctl list-units --type=service --state=running --output=json`
- Remote servers queried via `ssh user@host`
- `filter` and `exclude` use substring matching

### Kubernetes

- Creates one server per namespace (type: `cluster`)
- Deduplicates pods by `app` label (deployment replicas → single node)
- Maps service ports and ingress hosts

### Proxmox VE

- One server per PVE node (type: `hypervisor`)
- VMs: `shape: rectangle`, LXC containers: `shape: hexagon`
- Only running VMs/containers included
- Requires a token with at least `PVEAuditor` permissions

### Portainer

- Assigns all containers to the specified `server` hostname
- Uses `com.docker.compose.project` label for categorization
- Requires an API key from User Settings → Access tokens

## Development

```bash
make build      # Compile the binary
make test       # Run all tests
make lint       # Run golangci-lint
make demo       # Generate demo.d2 from bundled testdata
make demo-svg   # Generate + render to SVG
make clean      # Remove build artifacts
```

Without make:

```bash
go build -o inframap-d2 .
go test ./...
go run . generate -c testdata/demo.yml -o demo.d2
```

## Contributing

See [HOW_TO_CONTRIBUTE.md](HOW_TO_CONTRIBUTE.md) for guidelines, architecture overview, and how to add new collectors.

## License

[MIT](LICENSE)
