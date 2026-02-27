package collector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ThomasCrouzet/inframap-d2/internal/model"
)

func init() {
	Register(func() RegisteredCollector { return &ProxmoxCollector{} })
}

// ProxmoxCollector collects VMs and containers from Proxmox VE via its API.
type ProxmoxCollector struct {
	APIURL   string
	TokenID  string
	Token    string
	Insecure bool
	// TestData paths for testing (bypasses HTTP calls)
	TestNodes     string
	TestResources string
}

func (pc *ProxmoxCollector) Metadata() CollectorMetadata {
	return CollectorMetadata{
		Name:        "proxmox",
		DisplayName: "Proxmox VE",
		Description: "Collects VMs and LXC containers from Proxmox VE clusters",
		ConfigKey:   "proxmox",
		DetectHint:  "",
	}
}

func (pc *ProxmoxCollector) Enabled(sources map[string]any) bool {
	section, ok := sources["proxmox"].(map[string]any)
	if !ok {
		return false
	}
	url, _ := section["api_url"].(string)
	return url != ""
}

func (pc *ProxmoxCollector) Configure(section map[string]any) error {
	if section == nil {
		return nil
	}
	if v, ok := section["api_url"].(string); ok {
		pc.APIURL = v
	}
	if v, ok := section["token_id"].(string); ok {
		pc.TokenID = v
	}
	if v, ok := section["token"].(string); ok {
		pc.Token = v
	}
	if pc.TokenID == "" {
		pc.TokenID = os.Getenv("INFRAMAP_PROXMOX_TOKEN_ID")
	}
	if pc.Token == "" {
		pc.Token = os.Getenv("INFRAMAP_PROXMOX_TOKEN")
	}
	if v, ok := section["insecure"].(bool); ok {
		pc.Insecure = v
	}
	return nil
}

func (pc *ProxmoxCollector) Validate() []ValidationError {
	var errs []ValidationError
	if pc.APIURL == "" {
		errs = append(errs, ValidationError{
			Field:      "sources.proxmox.api_url",
			Message:    "api_url is required",
			Suggestion: "set the URL of your Proxmox VE instance, e.g. https://pve.local:8006",
		})
	}
	if pc.TokenID == "" || pc.Token == "" {
		errs = append(errs, ValidationError{
			Field:      "sources.proxmox.token_id",
			Message:    "token_id and token are required for API authentication",
			Suggestion: "create an API token in Proxmox: Datacenter → Permissions → API Tokens",
		})
	}
	return errs
}

type pveNodeResponse struct {
	Data []pveNode `json:"data"`
}

type pveResourceResponse struct {
	Data []pveResource `json:"data"`
}

type pveNode struct {
	Node   string  `json:"node"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu"`
	MaxCPU int     `json:"maxcpu"`
	Mem    int64   `json:"mem"`
	MaxMem int64   `json:"maxmem"`
}

type pveResource struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"` // "qemu" or "lxc"
	Node   string  `json:"node"`
	VMID   int     `json:"vmid"`
	Name   string  `json:"name"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu"`
	MaxCPU int     `json:"maxcpu"`
	Mem    int64   `json:"mem"`
	MaxMem int64   `json:"maxmem"`
}

func (pc *ProxmoxCollector) Collect(infra *model.Infrastructure) error {
	nodes, err := pc.getNodes()
	if err != nil {
		return fmt.Errorf("getting nodes: %w", err)
	}

	resources, err := pc.getResources()
	if err != nil {
		return fmt.Errorf("getting resources: %w", err)
	}

	// Create a server for each PVE node
	for _, node := range nodes {
		serverName := node.Node
		server, exists := infra.Servers[serverName]
		if !exists {
			server = &model.Server{
				Hostname: serverName,
				Label:    serverName,
				Type:     model.ServerTypeHypervisor,
				Online:   node.Status == "online",
			}
			infra.Servers[serverName] = server
		} else {
			server.Type = model.ServerTypeHypervisor
		}
	}

	// Group resources by node
	for _, res := range resources {
		if res.Status != "running" {
			continue
		}

		server, exists := infra.Servers[res.Node]
		if !exists {
			continue
		}

		svcType := model.ServiceTypeVM
		if res.Type == "lxc" {
			svcType = model.ServiceTypeLXC
		}

		svc := &model.Service{
			Name:     res.Name,
			Type:     svcType,
			Category: "virtualization",
		}

		server.AddService(svc)
	}

	return nil
}

func (pc *ProxmoxCollector) getNodes() ([]pveNode, error) {
	if pc.TestNodes != "" {
		return loadPVENodes(pc.TestNodes)
	}
	return pc.apiGetNodes("/api2/json/nodes")
}

func (pc *ProxmoxCollector) getResources() ([]pveResource, error) {
	if pc.TestResources != "" {
		return loadPVEResources(pc.TestResources)
	}
	return pc.apiGetResources("/api2/json/cluster/resources?type=vm")
}

func (pc *ProxmoxCollector) httpClient() *http.Client {
	client := &http.Client{Timeout: 30 * time.Second}
	if pc.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // user-configured
		}
	}
	return client
}

func (pc *ProxmoxCollector) apiGetNodes(path string) ([]pveNode, error) {
	body, err := pc.apiRequest(path)
	if err != nil {
		return nil, err
	}
	var resp pveNodeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (pc *ProxmoxCollector) apiGetResources(path string) ([]pveResource, error) {
	body, err := pc.apiRequest(path)
	if err != nil {
		return nil, err
	}
	var resp pveResourceResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (pc *ProxmoxCollector) apiRequest(path string) ([]byte, error) {
	client := pc.httpClient()

	req, err := http.NewRequest("GET", pc.APIURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", pc.TokenID, pc.Token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("proxmox API returned %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func loadPVENodes(path string) ([]pveNode, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var resp pveNodeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func loadPVEResources(path string) ([]pveResource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var resp pveResourceResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
