package internal

import (
	"os"
	"path/filepath"
)

func UserConfigDir() string {
	if v := os.Getenv("XDG_CONFIG_DIR"); v != "" {
		return v
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDir, ".config")
}
