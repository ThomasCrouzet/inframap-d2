package model

// Device represents a Tailscale peer that is not a server (phone, laptop, IoT).
type Device struct {
	Hostname    string
	OS          string
	TailscaleIP string
	Online      bool
	Tags        []string
}
