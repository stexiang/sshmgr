package sshutil

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
)

// HostKeyFingerprint fetches SSH host key and returns SHA256 fingerprint.
func HostKeyFingerprint(ip string) (string, error) {
	cmd := exec.Command(
		"ssh-keyscan",
		"-T", "2",
		"-t", "ed25519,ecdsa,rsa",
		ip,
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}

		rawKey, err := base64.StdEncoding.DecodeString(fields[2])
		if err != nil {
			continue
		}

		sum := sha256.Sum256(rawKey)
		fp := base64.StdEncoding.EncodeToString(sum[:])
		return "SHA256:" + fp, nil
	}

	return "", fmt.Errorf("no ssh hostkey found")
}
