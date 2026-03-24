package tui

import (
	"github.com/gather-system/gst-agent-launcher/config"
	"github.com/gather-system/gst-agent-launcher/launcher"
)

// configLoadedMsg is sent when the config has been loaded.
type configLoadedMsg struct {
	config *config.Config
}

// configReloadedMsg is sent when the config file has been modified and reloaded.
type configReloadedMsg struct {
	config *config.Config
}

// launchResultMsg is sent when the launch completes.
type launchResultMsg struct {
	result *launcher.LaunchResult
}

// launchErrMsg is sent when the launch fails.
type launchErrMsg struct {
	err error
}

// errMsg is sent when an error occurs.
type errMsg struct {
	err error
}

// monitorResultMsg is sent when the monitor-only launch completes.
type monitorResultMsg struct{ err error }

// toastMsg clears the toast after timeout.
type toastMsg struct{ id int }

// processScanMsg is sent when process detection completes.
type processScanMsg struct {
	running map[int]bool // keyed by agent index
	err     error
}
