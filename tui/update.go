package tui

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/gather-system/gst-agent-launcher/config"
	gitpkg "github.com/gather-system/gst-agent-launcher/git"
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
		m.projectNames = buildProjectNames(m.config)
		m.gitStatuses = make(map[int]gitpkg.RepoStatus)
		m.gitLoading = true
		// Start config file watcher and git status check.
		return m, tea.Batch(startConfigWatcher(), gitStatusCmd(m.config.Agents, m.pathValid))

	case configReloadedMsg:
		// Preserve current selection by name.
		selectedNames := m.selectedAgentNames()
		m.config = msg.config
		m.items = buildItems(m.config)
		m.pathValid = make(map[int]bool)
		for i, agent := range m.config.Agents {
			_, err := os.Stat(agent.Path)
			m.pathValid[i] = err == nil
		}
		m.projectNames = buildProjectNames(m.config)
		// Restore selection by name matching.
		m.selected = make(map[int]bool)
		nameSet := make(map[string]bool)
		for _, n := range selectedNames {
			nameSet[n] = true
		}
		for i, agent := range m.config.Agents {
			if nameSet[agent.Name] {
				m.selected[i] = true
			}
		}
		// Adjust cursor.
		if m.cursor >= len(m.items) {
			m.cursor = firstSelectableIndex(m.items)
		}
		return m, tea.Batch(setToast(&m, "設定檔已重新載入"), waitForConfigReload())

	case gitStatusMsg:
		m.gitLoading = false
		m.gitStatuses = make(map[int]gitpkg.RepoStatus)
		for _, s := range msg.statuses {
			if s.Error == nil {
				m.gitStatuses[s.AgentIndex] = s
			}
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

	case tea.MouseClickMsg:
		if m.view == viewList && msg.Button == tea.MouseLeft {
			idx := m.itemAtY(msg.Y)
			if idx >= 0 && !m.items[idx].isGroup {
				now := time.Now().UnixMilli()
				if idx == m.lastClickY && now-m.lastClickTime < 350 {
					// Double-click: select only this agent and launch
					m.selected = make(map[int]bool)
					m.selected[m.items[idx].index] = true
					m.lastClickY = -1
					m.lastClickTime = 0
					m.view = viewConfirm
					return m, nil
				}
				m.cursor = idx
				m.toggleCurrent()
				m.lastClickY = idx
				m.lastClickTime = now
			}
		}
		return m, nil

	case tea.MouseWheelMsg:
		if m.view == viewList {
			if msg.Button == tea.MouseWheelUp {
				m.moveCursorUp()
			} else if msg.Button == tea.MouseWheelDown {
				m.moveCursorDown()
			}
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
		case viewProject:
			return m.updateProject(msg)
		}
	}

	return m, nil
}

// updateList handles key presses in the list view.
func (m Model) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.searchMode {
		return m.updateSearch(msg)
	}

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

	case "e":
		path := config.ConfigPath()
		if path != "" {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "notepad"
			}
			c := exec.Command(editor, path)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				if err != nil {
					return errMsg{err}
				}
				return nil
			})
		}

	case "?":
		m.view = viewHelp

	case "P":
		if len(m.projectNames) > 0 {
			m.view = viewProject
			m.projectCursor = 0
		}

	case "/":
		m.searchMode = true
		m.searchQuery = ""

	case "M":
		if m.config != nil && m.config.Monitor.Command != "" {
			return m, m.doLaunchMonitorOnly()
		}
	}

	return m, cmd
}

// updateSearch handles key presses in search mode.
func (m Model) updateSearch(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "escape":
		m.searchMode = false
		m.searchQuery = ""
		m.cursor = firstSelectableIndex(m.items)
	case "enter":
		m.searchMode = false
	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.adjustCursorForSearch()
		}
	case "up", "k":
		m.moveCursorUp()
	case "down", "j":
		m.moveCursorDown()
	case "space", " ":
		m.toggleCurrent()
	case "ctrl+c":
		return m, tea.Quit
	default:
		if len(key) == 1 {
			m.searchQuery += key
			m.adjustCursorForSearch()
		}
	}
	return m, nil
}

// updateProject handles key presses in the project selection view.
func (m Model) updateProject(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "q":
		m.view = viewList
	case "up", "k":
		if m.projectCursor > 0 {
			m.projectCursor--
		}
	case "down", "j":
		if m.projectCursor < len(m.projectNames)-1 {
			m.projectCursor++
		}
	case "enter":
		name := m.projectNames[m.projectCursor]
		count := m.selectProject(name)
		m.view = viewList
		return m, setToast(&m, fmt.Sprintf("已選取 %s (%d agents)", name, count))
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
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


// moveCursorUp moves the cursor up, skipping group headers and search-hidden items.
func (m *Model) moveCursorUp() {
	for i := m.cursor - 1; i >= 0; i-- {
		if !m.items[i].isGroup && m.matchesSearch(m.items[i]) {
			m.cursor = i
			return
		}
	}
}

// moveCursorDown moves the cursor down, skipping group headers and search-hidden items.
func (m *Model) moveCursorDown() {
	for i := m.cursor + 1; i < len(m.items); i++ {
		if !m.items[i].isGroup && m.matchesSearch(m.items[i]) {
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
