package sshutil

import (
	"fmt"
	"os/exec"
	"strings"
)

// RemoteHostname attempts to execute 'hostname' via ssh.
// Requires user to have password/key ready; otherwise returns error.
func RemoteHostname(user, ip string) (string, error) {
	out, err := exec.Command(
		"ssh",
		"-o", "BatchMode=yes",
		fmt.Sprintf("%s@%s", user, ip),
		"hostname",
	).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
