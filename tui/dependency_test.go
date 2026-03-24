package tui

import (
	"testing"

	"github.com/gather-system/gst-agent-launcher/config"
)

func TestCheckDependencies_NoDeps(t *testing.T) {
	m := Model{
		config: &config.Config{
			Agents: []config.Agent{
				{Name: "PM", Path: "/pm", Group: "PM"},
				{Name: "Core", Path: "/core", Group: "Core"},
			},
		},
		selected:      map[int]bool{0: true},
		runningAgents: map[int]bool{},
	}

	unmet := m.checkDependencies()
	if len(unmet) != 0 {
		t.Errorf("expected no unmet deps, got %v", unmet)
	}
}

func TestCheckDependencies_UnmetDeps(t *testing.T) {
	m := Model{
		config: &config.Config{
			Agents: []config.Agent{
				{Name: "GST-process-vision", Path: "/core-pv", Group: "Core"},
				{Name: "ProcessVision-App", Path: "/app-pv", Group: "App",
					Dependencies: []string{"GST-process-vision"}},
			},
		},
		selected:      map[int]bool{1: true}, // only App selected
		runningAgents: map[int]bool{},
	}

	unmet := m.checkDependencies()
	if len(unmet) != 1 || unmet[0] != "GST-process-vision" {
		t.Errorf("expected [GST-process-vision], got %v", unmet)
	}
}

func TestCheckDependencies_DepSelected(t *testing.T) {
	m := Model{
		config: &config.Config{
			Agents: []config.Agent{
				{Name: "GST-process-vision", Path: "/core-pv", Group: "Core"},
				{Name: "ProcessVision-App", Path: "/app-pv", Group: "App",
					Dependencies: []string{"GST-process-vision"}},
			},
		},
		selected:      map[int]bool{0: true, 1: true}, // both selected
		runningAgents: map[int]bool{},
	}

	unmet := m.checkDependencies()
	if len(unmet) != 0 {
		t.Errorf("expected no unmet deps when dep is selected, got %v", unmet)
	}
}

func TestCheckDependencies_DepRunning(t *testing.T) {
	m := Model{
		config: &config.Config{
			Agents: []config.Agent{
				{Name: "GST-process-vision", Path: "/core-pv", Group: "Core"},
				{Name: "ProcessVision-App", Path: "/app-pv", Group: "App",
					Dependencies: []string{"GST-process-vision"}},
			},
		},
		selected:      map[int]bool{1: true}, // only App selected
		runningAgents: map[int]bool{0: true},  // dep is running
	}

	unmet := m.checkDependencies()
	if len(unmet) != 0 {
		t.Errorf("expected no unmet deps when dep is running, got %v", unmet)
	}
}

func TestCheckDependencies_MultipleDeps(t *testing.T) {
	m := Model{
		config: &config.Config{
			Agents: []config.Agent{
				{Name: "GST-smb", Path: "/smb", Group: "Core"},
				{Name: "GST-fda-part11", Path: "/fda", Group: "Core"},
				{Name: "SMB-App", Path: "/app", Group: "App",
					Dependencies: []string{"GST-smb", "GST-fda-part11"}},
			},
		},
		selected:      map[int]bool{2: true},
		runningAgents: map[int]bool{0: true}, // GST-smb running, GST-fda-part11 not
	}

	unmet := m.checkDependencies()
	if len(unmet) != 1 || unmet[0] != "GST-fda-part11" {
		t.Errorf("expected [GST-fda-part11], got %v", unmet)
	}
}

func TestSelectDependencies(t *testing.T) {
	m := Model{
		config: &config.Config{
			Agents: []config.Agent{
				{Name: "GST-smb", Path: "/smb", Group: "Core"},
				{Name: "GST-fda-part11", Path: "/fda", Group: "Core"},
				{Name: "SMB-App", Path: "/app", Group: "App"},
			},
		},
		selected: map[int]bool{2: true},
	}

	m.selectDependencies([]string{"GST-smb", "GST-fda-part11"})

	if !m.selected[0] || !m.selected[1] {
		t.Errorf("expected deps to be selected, got %v", m.selected)
	}
}
