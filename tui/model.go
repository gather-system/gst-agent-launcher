package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/gather-system/gst-agent-launcher/config"
)

// Model is the Bubble Tea model for the launcher TUI.
type Model struct {
	config *config.Config
	err    error
}

// NewModel creates a new Model with default state.
func NewModel() Model {
	return Model{}
}

// Init loads the configuration on startup.
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			return errMsg{err}
		}
		return configLoadedMsg{cfg}
	}
}

// configLoadedMsg is sent when the config has been loaded.
type configLoadedMsg struct {
	config *config.Config
}

// errMsg is sent when an error occurs.
type errMsg struct {
	err error
}
