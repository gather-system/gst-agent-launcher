package tui

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/gather-system/gst-agent-launcher/config"
	"github.com/gather-system/gst-agent-launcher/launcher"
)

// Update handles incoming messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case configLoadedMsg:
		m.config = msg.config
		m.items = buildItems(m.config)
		m.cursor = firstSelectableIndex(m.items)
		m.monitorOn = m.config.Monitor.Enabled
		m.pathValid = make(map[int]bool)
		for i, agent := range m.config.Agents {
			_, err := os.Stat(agent.Path)
			m.pathValid[i] = err == nil
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case launchResultMsg:
		m.result = msg.result
		m.view = viewResult
		for _, name := range msg.result.Launched {
			if name == "Monitor" {
				m.monitorLaunched = true
				break
			}
		}
		// Save session for r-key restore.
		config.SaveSession(m.selectedAgentNames(), m.monitorOn)
		return m, nil

	case launchErrMsg:
		m.err = msg.err
		m.view = viewList
		return m, nil

	case toastMsg:
		if msg.id == m.toastTimer {
			m.toast = ""
		}
		return m, nil

	case monitorResultMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.monitorLaunched = true
			return m, setToast(&m, "Monitor 已啟動")
		}
		return m, nil

	case tea.KeyPressMsg:
		switch m.view {
		case viewList:
			return m.updateList(msg)
		case viewConfirm:
			return m.updateConfirm(msg)
		case viewResult:
			return m.updateResult(msg)
		case viewHelp:
			m.view = viewList
			return m, nil
		}
	}

	return m, nil
}

// updateList handles key presses in the list view.
func (m Model) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		m.moveCursorUp()

	case "down", "j":
		m.moveCursorDown()

	case "space", " ":
		m.toggleCurrent()

	case "enter":
		if m.selectedCount() > 0 || m.monitorOn {
			m.view = viewConfirm
		} else {
			return m, setToast(&m, "請先選擇至少一個 Agent")
		}

	case "escape":
		m.resetSelection()
		m.monitorOn = false
		return m, setToast(&m, "已清除所有選擇")

	case "a":
		m.toggleAll()
		count := m.selectedCount()
		if count > 0 {
			cmd = setToast(&m, fmt.Sprintf("已勾選全部 (%d)", count))
		} else {
			cmd = setToast(&m, "已取消全部勾選")
		}

	case "c":
		m.toggleGroup("Core")
		cmd = groupToast(&m, "Core")

	case "p":
		m.toggleGroup("PM")
		cmd = groupToast(&m, "PM")

	case "o":
		m.toggleGroup("App")
		cmd = groupToast(&m, "App")

	case "l":
		m.toggleGroup("Leyu")
		cmd = groupToast(&m, "Leyu")

	case "m":
		m.monitorOn = !m.monitorOn
		if m.monitorOn {
			cmd = setToast(&m, "Monitor: ON")
		} else {
			cmd = setToast(&m, "Monitor: OFF")
		}

	case "r":
		session, err := config.LoadSession()
		if err != nil {
			return m, setToast(&m, "沒有上次的選擇紀錄")
		}
		m.restoreSession(session)
		count := m.selectedCount()
		return m, setToast(&m, fmt.Sprintf("已恢復上次選擇 (%d)", count))

	case "?":
		m.view = viewHelp

	case "M":
		if m.config != nil && m.config.Monitor.Command != "" {
			return m, m.doLaunchMonitorOnly()
		}
	}

	return m, cmd
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

// updateResult handles key presses in the result view.
func (m Model) updateResult(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	default:
		m.resetSelection()
		m.view = viewList
	}
	return m, nil
}

// updateConfirm handles key presses in the confirm view.
func (m Model) updateConfirm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		return m, m.doLaunch()
	case "n", "escape", "q":
		m.view = viewList
	}
	return m, nil
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

// moveCursorUp moves the cursor up, skipping group headers.
func (m *Model) moveCursorUp() {
	for i := m.cursor - 1; i >= 0; i-- {
		if !m.items[i].isGroup {
			m.cursor = i
			return
		}
	}
}

// moveCursorDown moves the cursor down, skipping group headers.
func (m *Model) moveCursorDown() {
	for i := m.cursor + 1; i < len(m.items); i++ {
		if !m.items[i].isGroup {
			m.cursor = i
			return
		}
	}
}

// toggleCurrent toggles the selection of the item at the cursor.
func (m *Model) toggleCurrent() {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return
	}
	item := m.items[m.cursor]
	if item.isGroup {
		return
	}
	m.selected[item.index] = !m.selected[item.index]
}

// toggleAll selects all agents if any are unselected, otherwise deselects all.
func (m *Model) toggleAll() {
	if m.config == nil {
		return
	}
	allSelected := true
	for i := range m.config.Agents {
		if !m.selected[i] {
			allSelected = false
			break
		}
	}
	for i := range m.config.Agents {
		m.selected[i] = !allSelected
	}
}

// toggleGroup selects all agents in a group if any are unselected, otherwise deselects all.
func (m *Model) toggleGroup(group string) {
	if m.config == nil {
		return
	}
	var indices []int
	for i, agent := range m.config.Agents {
		if agent.Group == group {
			indices = append(indices, i)
		}
	}
	allSelected := true
	for _, i := range indices {
		if !m.selected[i] {
			allSelected = false
			break
		}
	}
	for _, i := range indices {
		m.selected[i] = !allSelected
	}
}
