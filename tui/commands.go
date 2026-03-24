package tui

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/gather-system/gst-agent-launcher/config"
	gitpkg "github.com/gather-system/gst-agent-launcher/git"
	"github.com/gather-system/gst-agent-launcher/launcher"
)

// configWatchCh holds the channel from the config file watcher.
var configWatchCh <-chan *config.Config

// startConfigWatcher starts watching the config file and returns a command to wait for the first reload.
func startConfigWatcher() tea.Cmd {
	path := config.ConfigPath()
	if path == "" {
		return nil
	}
	ch, err := config.WatchConfig(path)
	if err != nil {
		return nil
	}
	configWatchCh = ch
	return waitForConfigReload()
}

// waitForConfigReload returns a command that waits for the next config reload event.
func waitForConfigReload() tea.Cmd {
	if configWatchCh == nil {
		return nil
	}
	return func() tea.Msg {
		cfg := <-configWatchCh
		return configReloadedMsg{cfg}
	}
}

// setToast sets a toast message and returns a command to clear it after 2 seconds.
func setToast(m *Model, msg string) tea.Cmd {
	m.toastTimer++
	m.toast = msg
	id := m.toastTimer
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return toastMsg{id}
	})
}

// groupToast generates a toast message for a group toggle.
func groupToast(m *Model, group string) tea.Cmd {
	sel, total := m.groupCount(group)
	if sel > 0 {
		return setToast(m, fmt.Sprintf("已勾選 %s 群組 (%d/%d)", group, sel, total))
	}
	return setToast(m, fmt.Sprintf("已取消 %s 群組", group))
}

// gitStatusCmd fetches git status for all agents with valid paths asynchronously.
func gitStatusCmd(agents []config.Agent, pathValid map[int]bool) tea.Cmd {
	return func() tea.Msg {
		runner := gitpkg.NewRunner()
		isGitRepo := func(i int) bool { return pathValid[i] }
		statuses := gitpkg.GetAllStatuses(context.Background(), runner, agents, isGitRepo)
		return gitStatusMsg{statuses: statuses}
	}
}

// doLaunch creates a command that performs the actual launch.
func (m Model) doLaunch() tea.Cmd {
	agents := m.selectedAgents()
	var monitor *config.Monitor
	// Skip monitor if already launched to avoid duplicate tabs.
	if m.monitorOn && !m.monitorLaunched && m.config != nil {
		monitor = &m.config.Monitor
	}

	return func() tea.Msg {
		result, err := launcher.LaunchAll(agents, monitor)
		if err != nil {
			return launchErrMsg{err}
		}
		return launchResultMsg{result}
	}
}

// doLaunchMonitorOnly launches only the monitor tab.
func (m Model) doLaunchMonitorOnly() tea.Cmd {
	monitor := m.config.Monitor
	return func() tea.Msg {
		err := launcher.LaunchMonitor(monitor)
		return monitorResultMsg{err}
	}
}
