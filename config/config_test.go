package config

import (
	"encoding/json"
	"testing"
)

func TestAgentWithDependencies(t *testing.T) {
	input := `{
		"agents": [
			{"name": "App", "path": "/app", "group": "App", "dependencies": ["Core-A", "Core-B"]}
		],
		"monitor": {"enabled": false, "command": ""}
	}`

	var cfg Config
	if err := json.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(cfg.Agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(cfg.Agents))
	}

	agent := cfg.Agents[0]
	if len(agent.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(agent.Dependencies))
	}
	if agent.Dependencies[0] != "Core-A" || agent.Dependencies[1] != "Core-B" {
		t.Errorf("unexpected dependencies: %v", agent.Dependencies)
	}
}

func TestAgentWithoutDependencies(t *testing.T) {
	input := `{
		"agents": [
			{"name": "Core", "path": "/core", "group": "Core"}
		],
		"monitor": {"enabled": false, "command": ""}
	}`

	var cfg Config
	if err := json.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	agent := cfg.Agents[0]
	if agent.Dependencies != nil {
		t.Errorf("expected nil dependencies, got %v", agent.Dependencies)
	}
}

func TestAgentEmptyDependencies(t *testing.T) {
	input := `{
		"agents": [
			{"name": "Core", "path": "/core", "group": "Core", "dependencies": []}
		],
		"monitor": {"enabled": false, "command": ""}
	}`

	var cfg Config
	if err := json.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	agent := cfg.Agents[0]
	if len(agent.Dependencies) != 0 {
		t.Errorf("expected empty dependencies, got %v", agent.Dependencies)
	}
}

func TestAgentMarshalOmitsEmptyDependencies(t *testing.T) {
	agent := Agent{Name: "Core", Path: "/core", Group: "Core"}
	data, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var m map[string]interface{}
	json.Unmarshal(data, &m)
	if _, exists := m["dependencies"]; exists {
		t.Error("expected dependencies to be omitted when empty")
	}
}

func TestLoadDefault(t *testing.T) {
	cfg, err := loadDefault()
	if err != nil {
		t.Fatalf("loadDefault failed: %v", err)
	}

	if len(cfg.Agents) == 0 {
		t.Fatal("expected agents in default config")
	}

	// Verify ProcessVision-App has dependencies.
	var pvApp *Agent
	for i, a := range cfg.Agents {
		if a.Name == "ProcessVision-App" {
			pvApp = &cfg.Agents[i]
			break
		}
	}
	if pvApp == nil {
		t.Fatal("ProcessVision-App not found in default config")
	}
	if len(pvApp.Dependencies) != 1 || pvApp.Dependencies[0] != "GST-process-vision" {
		t.Errorf("unexpected ProcessVision-App dependencies: %v", pvApp.Dependencies)
	}

	// Verify agents without dependencies still parse correctly.
	var pm *Agent
	for i, a := range cfg.Agents {
		if a.Name == "PM" {
			pm = &cfg.Agents[i]
			break
		}
	}
	if pm == nil {
		t.Fatal("PM not found in default config")
	}
	if pm.Dependencies != nil {
		t.Errorf("expected PM to have nil dependencies, got %v", pm.Dependencies)
	}
}
