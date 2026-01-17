package sys

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func service(name string) string { return "sshmgr:" + name }

func KeychainSet(name, user, password string) error {
	cmd := exec.Command("security", "add-generic-password",
		"-a", user,
		"-s", service(name),
		"-w", password,
		"-U",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("keychain set failed: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func KeychainGet(name, user string) (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-a", user,
		"-s", service(name),
		"-w",
	)
	var b bytes.Buffer
	cmd.Stdout = &b
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("keychain get failed: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func KeychainDelete(name, user string) error {
	cmd := exec.Command("security", "delete-generic-password",
		"-a", user,
		"-s", service(name),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("keychain delete failed: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
