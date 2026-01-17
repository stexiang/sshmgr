package app

import (
	"os"
	"path/filepath"
)

func DefaultDBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sshmgr", "sshmgr.db")
}

func EnsureParentDir(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0o755)
}
