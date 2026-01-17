package netx

import (
	"context"
	"net"
)

func PickOneIP(ips []net.IP) string {
	// 1) prefer non-loopback, non-linklocal IPv4
	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			if ip4.IsLoopback() || ip4.IsLinkLocalUnicast() {
				continue
			}
			return ip4.String()
		}
	}
	// 2) fallback: any IPv4 (even loopback)
	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			return ip4.String()
		}
	}
	// 3) fallback: any non-loopback IPv6
	for _, ip := range ips {
		if ip.To4() == nil && len(ip) == net.IPv6len {
			if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			return ip.String()
		}
	}
	// 4) last resort
	if len(ips) > 0 {
		return ips[0].String()
	}
	return ""
}

func ResolveHost(ctx context.Context, host string) (string, error) {
	addrs, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return "", err
	}
	return PickOneIP(addrs), nil
}
