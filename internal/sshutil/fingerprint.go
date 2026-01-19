package sshutil

import (
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
)

// Fingerprint fetches SSH host key and returns SHA256 hash of the raw keyscan output.
func Fingerprint(ip string) (string, error) {
	out, err := exec.Command("ssh-keyscan", "-T", "3", ip).Output()
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(out)
	return hex.EncodeToString(sum[:]), nil
}
