package model

import (
	"fmt"
	"strconv"
	"strings"
)

// PortMapping represents a port binding.
type PortMapping struct {
	HostIP        string
	HostPort      int
	ContainerPort int
	Protocol      string // tcp or udp
}

// String returns a human-readable port mapping.
func (p PortMapping) String() string {
	proto := p.Protocol
	if proto == "" || proto == "tcp" {
		proto = ""
	} else {
		proto = "/" + proto
	}
	if p.HostPort == p.ContainerPort {
		return fmt.Sprintf("%d%s", p.HostPort, proto)
	}
	return fmt.Sprintf("%d→%d%s", p.HostPort, p.ContainerPort, proto)
}

// ParsePortMapping parses a Docker port string like "8080:80" or "127.0.0.1:8080:80/tcp".
func ParsePortMapping(s string) PortMapping {
	pm := PortMapping{Protocol: "tcp"}

	// Split protocol
	if idx := strings.Index(s, "/"); idx != -1 {
		pm.Protocol = s[idx+1:]
		s = s[:idx]
	}

	parts := strings.Split(s, ":")
	switch len(parts) {
	case 1:
		port, _ := strconv.Atoi(parts[0])
		pm.HostPort = port
		pm.ContainerPort = port
	case 2:
		pm.HostPort, _ = strconv.Atoi(parts[0])
		pm.ContainerPort, _ = strconv.Atoi(parts[1])
	case 3:
		pm.HostIP = parts[0]
		pm.HostPort, _ = strconv.Atoi(parts[1])
		pm.ContainerPort, _ = strconv.Atoi(parts[2])
	}
	return pm
}
