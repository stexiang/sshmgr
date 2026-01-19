package netutil

import (
	"net"
	"strings"
)

func ReverseLookup(ip string) (string, error) {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return "", err
	}
	// strip trailing dot
	return strings.TrimSuffix(names[0], "."), nil
}
