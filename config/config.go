package config

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
)

//go:embed default.json
var defaultConfigFS embed.FS

// Agent represents a single agent entry in the configuration.
type Agent struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Group string `json:"group"`
}

// Monitor represents the monitor configuration.
type Monitor struct {
	Enabled bool   `json:"enabled"`
	Command string `json:"command"`
}

// Project represents a predefined agent selection preset.
type Project struct {
	Agents      []string `json:"agents"`
	Description string   `json:"description"`
}

// Config represents the top-level configuration.
type Config struct {
	Agents   []Agent            `json:"agents"`
	Monitor  Monitor            `json:"monitor"`
	Projects map[string]Project `json:"projects,omitempty"`
}

// Load reads the configuration from the following locations (in priority order):
//  1. ~/.config/gst-launcher/agents.json
//  2. agents.json next to the executable
//  3. Embedded default.json
func Load() (*Config, error) {
	// 1. User config directory
	home, err := os.UserHomeDir()
	if err == nil {
		userPath := filepath.Join(home, ".config", "gst-launcher", "agents.json")
		if cfg, err := loadFromFile(userPath); err == nil {
			return cfg, nil
		}
	}

	// 2. Executable directory
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		localPath := filepath.Join(exeDir, "agents.json")
		if cfg, err := loadFromFile(localPath); err == nil {
			return cfg, nil
		}
	}

	// 3. Embedded default
	return loadDefault()
}

func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// EnsureUserConfig creates ~/.config/gst-launcher/agents.json from the
// embedded default if it does not already exist.
func EnsureUserConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	dir := filepath.Join(home, ".config", "gst-launcher")
	target := filepath.Join(dir, "agents.json")

	if _, err := os.Stat(target); err == nil {
		return // already exists
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}

	data, err := defaultConfigFS.ReadFile("default.json")
	if err != nil {
		return
	}

	os.WriteFile(target, data, 0o644)
}

func loadDefault() (*Config, error) {
	data, err := defaultConfigFS.ReadFile("default.json")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
