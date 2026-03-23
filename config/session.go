package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Session represents the last launch session.
type Session struct {
	Agents    []string `json:"agents"`
	MonitorOn bool     `json:"monitorOn"`
}

// sessionPath returns the path to last-session.json.
func sessionPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gst-launcher", "last-session.json"), nil
}

// SaveSession saves the current selection to last-session.json.
func SaveSession(agents []string, monitorOn bool) error {
	path, err := sessionPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	session := Session{
		Agents:    agents,
		MonitorOn: monitorOn,
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// LoadSession loads the last session from last-session.json.
func LoadSession() (*Session, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}
