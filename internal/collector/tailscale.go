package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
)

func init() {
	Register(func() RegisteredCollector { return &TailscaleCollector{} })
}

// TailscaleCollector parses `tailscale status --json` output.
type TailscaleCollector struct {
	JsonFile       string
	IncludeOffline bool
}

func (tc *TailscaleCollector) Metadata() CollectorMetadata {
	return CollectorMetadata{
		Name:        "tailscale",
		DisplayName: "Tailscale",
		Description: "Collects Tailscale VPN peers, IPs, and online status",
		ConfigKey:   "tailscale",
		DetectHint:  "tailscale",
	}
}

func (tc *TailscaleCollector) Enabled(sources map[string]any) bool {
	section, ok := sources["tailscale"].(map[string]any)
	if !ok {
		return false
	}
	if enabled, ok := section["enabled"].(bool); ok {
		return enabled
	}
	return false
}

func (tc *TailscaleCollector) Configure(section map[string]any) error {
	if section == nil {
		return nil
	}
	if v, ok := section["json_file"].(string); ok {
		tc.JsonFile = v
	}
	if v, ok := section["include_offline"].(bool); ok {
		tc.IncludeOffline = v
	}
	return nil
}

func (tc *TailscaleCollector) Validate() []ValidationError {
	var errs []ValidationError
	if tc.JsonFile != "" {
		if _, err := os.Stat(tc.JsonFile); err != nil {
			errs = append(errs, ValidationError{
				Field:      "sources.tailscale.json_file",
				Message:    fmt.Sprintf("file not found: %s", tc.JsonFile),
				Suggestion: "check the path or remove json_file to use live tailscale status",
			})
		}
	} else {
		// Check if tailscale binary is available
		if _, err := exec.LookPath("tailscale"); err != nil {
			errs = append(errs, ValidationError{
				Field:      "sources.tailscale",
				Message:    "tailscale binary not found in PATH",
				Suggestion: "install tailscale or provide a json_file path",
			})
		}
	}
	return errs
}

// tailscaleStatus represents the JSON output of `tailscale status --json`.
type tailscaleStatus struct {
	Self            tailscalePeer            `json:"Self"`
	Peer            map[string]tailscalePeer `json:"Peer"`
	MagicDNSSuffix  string                   `json:"MagicDNSSuffix"`
	CurrentTailnet  *tailscaleTailnet        `json:"CurrentTailnet"`
}

type tailscaleTailnet struct {
	Name string `json:"Name"`
}

type tailscalePeer struct {
	HostName     string   `json:"HostName"`
	DNSName      string   `json:"DNSName"`
	OS           string   `json:"OS"`
	TailscaleIPs []string `json:"TailscaleIPs"`
	Online       bool     `json:"Online"`
	Tags         []string `json:"Tags"`
}

func (tc *TailscaleCollector) Collect(infra *model.Infrastructure) error {
	data, err := tc.getData()
	if err != nil {
		return fmt.Errorf("getting tailscale data: %w", err)
	}

	var status tailscaleStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return fmt.Errorf("parsing tailscale json: %w", err)
	}

	// Set tailnet name
	if status.CurrentTailnet != nil {
		infra.TailnetName = status.CurrentTailnet.Name
	}

	// Process self
	tc.processPeer(infra, status.Self)

	// Process peers
	for _, peer := range status.Peer {
		if !peer.Online && !tc.IncludeOffline {
			continue
		}
		tc.processPeer(infra, peer)
	}

	return nil
}

func (tc *TailscaleCollector) getData() ([]byte, error) {
	if tc.JsonFile != "" {
		return os.ReadFile(tc.JsonFile)
	}

	cmd := exec.Command("tailscale", "status", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running tailscale status: %w", err)
	}
	return output, nil
}

func (tc *TailscaleCollector) processPeer(infra *model.Infrastructure, peer tailscalePeer) {
	hostname := strings.ToLower(peer.HostName)
	if hostname == "" {
		return
	}

	tsIP := ""
	if len(peer.TailscaleIPs) > 0 {
		tsIP = peer.TailscaleIPs[0]
	}

	// Check if this peer matches an existing server
	if server, exists := infra.Servers[hostname]; exists {
		server.TailscaleIP = tsIP
		server.OS = peer.OS
		server.Online = peer.Online
		return
	}

	// Check if it's a server-like device (tagged as server)
	isServer := false
	for _, tag := range peer.Tags {
		if strings.Contains(tag, "server") {
			isServer = true
			break
		}
	}

	if isServer {
		infra.Servers[hostname] = &model.Server{
			Hostname:    hostname,
			Label:       hostname,
			TailscaleIP: tsIP,
			OS:          peer.OS,
			Online:      peer.Online,
			Type:        model.ServerTypeLab,
		}
		return
	}

	// It's a device (phone, laptop, etc.)
	infra.Devices[hostname] = &model.Device{
		Hostname:    hostname,
		OS:          peer.OS,
		TailscaleIP: tsIP,
		Online:      peer.Online,
		Tags:        peer.Tags,
	}
}
