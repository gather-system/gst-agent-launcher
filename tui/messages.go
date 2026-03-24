package tui

import (
	"github.com/gather-system/gst-agent-launcher/config"
	gitpkg "github.com/gather-system/gst-agent-launcher/git"
	"github.com/gather-system/gst-agent-launcher/health"
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

// healthResultMsg is sent when health checks complete.
type healthResultMsg struct {
	results      []health.CheckResult
	gitAvailable bool
}

// gitStatusMsg is sent when git status checks complete.
type gitStatusMsg struct {
	statuses []gitpkg.RepoStatus
}

// processScanMsg is sent when process detection completes.
type processScanMsg struct {
	running map[int]bool
	err     error
}

// dashboardTickMsg triggers dashboard auto-refresh.
type dashboardTickMsg struct{ id int }

// batchCompleteMsg is sent when a batch git operation completes.
type batchCompleteMsg struct {
	results []gitpkg.BatchResult
}
