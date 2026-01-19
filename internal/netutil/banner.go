package netutil

import (
	"net"
	"strings"
	"time"
)

// SSHBanner dials port 22 and returns the first SSH handshake line.
func SSHBanner(ip string, timeout time.Duration) (string, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "22"), timeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

// ExtractHostname attempts to find "xxx.local" in banner text.
func ExtractHostname(banner string) string {
	fields := strings.Fields(banner)
	for _, f := range fields {
		if strings.HasSuffix(f, ".local") {
			return f
		}
	}
	return ""
}
