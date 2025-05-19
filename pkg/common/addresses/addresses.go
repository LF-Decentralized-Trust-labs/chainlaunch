package addresses

import (
	"fmt"
	"net"
	"os"
)

// GetExternalIP returns the external IP address of the node
func GetExternalIP() (string, error) {
	// Try to get external IP from environment variable first
	if externalIP := os.Getenv("EXTERNAL_IP"); externalIP != "" {
		return externalIP, nil
	}

	// Get local network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Look for a suitable non-loopback interface with an IPv4 address
	for _, iface := range interfaces {
		// Skip loopback, down interfaces, and interfaces without addresses
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Check if this is an IP network address
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// Skip loopback and IPv6 addresses
			ip := ipNet.IP.To4()
			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Skip link-local addresses
			if ip[0] == 169 && ip[1] == 254 {
				continue
			}

			// Found a suitable IP address
			return ip.String(), nil
		}
	}

	// Fallback to localhost if no suitable interface is found
	return "127.0.0.1", nil
}
