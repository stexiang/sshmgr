package netutil

import (
	"net"
	"strings"
	"time"
)

// SHBanner 连接目标 IP 的 22 端口，并读取返回的第一行 SSH 握手信息
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

// ExtractHostname 尝试从 SSH Banner 文本中提取形如 "xxx.local" 的主机名。
func ExtractHostname(banner string) string {
	fields := strings.Fields(banner)
	for _, f := range fields {
		if strings.HasSuffix(f, ".local") {
			return f
		}
	}
	return ""
}
