package utils

import (
	"fmt"
	"net"
	"time"
)

// IsPortAvailable checks if a port is available for use
func IsPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// FindAvailablePort finds an available port starting from the given port
// It will try the given port first, and if not available, it will try the next ports
// until it finds an available one or reaches maxPort
func FindAvailablePort(startPort, maxPort int) (int, error) {
	for port := startPort; port <= maxPort; port++ {
		if IsPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found between %d and %d", startPort, maxPort)
}

// WaitForPortToBecomeAvailable waits for a port to become available with timeout
func WaitForPortToBecomeAvailable(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if IsPortAvailable(port) {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}
