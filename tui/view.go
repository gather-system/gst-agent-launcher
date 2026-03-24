package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// View renders the TUI.
func (m Model) View() tea.View {
	var content string
	switch m.view {
	case viewConfirm:
		content = m.viewConfirm()
	case viewResult:
		content = m.viewResult()
	case viewHelp:
		content = m.viewHelpOverlay()
	case viewProject:
		content = m.viewProjectSelect()
	default:
		content = m.viewList()
	}
	v := tea.NewView(content)
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// viewList renders the agent selection list.
func (m Model) viewList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GST Agent Launcher"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("q: quit"))
		b.WriteString("\n")
		return b.String()
	}

	if m.config == nil {
		b.WriteString("Loading...\n")
		return b.String()
	}

	if len(m.items) == 0 {
		b.WriteString("No agents configured.\n")
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("q: quit"))
		b.WriteString("\n")
		return b.String()
	}

	// Monitor status
	if m.monitorLaunched {
		b.WriteString(successStyle.Render("Monitor [R]"))
		b.WriteString("\n\n")
	}

	// Render list items (filtered by search query)
	for i, item := range m.items {
		if item.isGroup {
			// Skip group header if no children match search
			hasMatch := false
			for j := i + 1; j < len(m.items) && !m.items[j].isGroup; j++ {
				if m.matchesSearch(m.items[j]) {
					hasMatch = true
					break
				}
			}
			if !hasMatch {
				continue
			}
			sel, tot := m.groupCount(item.group)
			style := groupStyle(item.group)
			b.WriteString(style.Render(fmt.Sprintf("── %s (%d/%d) ──", item.group, sel, tot)))
			b.WriteString("\n")
			continue
		}

		if !m.matchesSearch(item) {
			continue
		}

		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		check := "[ ]"
		if m.selected[item.index] {
			check = selectedStyle.Render("[x]")
		}

		name := item.agent.Name
		if !m.pathValid[item.index] {
			name = invalidStyle.Render(name + " [!]")
		} else if i == m.cursor {
			name = cursorStyle.Render(name)
		}
		if m.runningAgents[item.index] {
			name += " " + successStyle.Render("[R]")
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, name))
	}

	// Toast
	if m.toast != "" {
		b.WriteString(toastStyle.Render(m.toast))
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	var statusText string
	if m.searchMode {
		statusText = fmt.Sprintf("搜尋: %s_", m.searchQuery)
	} else {
		monitorStatus := "OFF"
		if m.monitorOn {
			monitorStatus = "ON"
		}
		statusText = fmt.Sprintf("已選: %d 個 Agent | Monitor: %s | Enter=啟動 q=退出",
			m.selectedCount(), monitorStatus)
	}
	b.WriteString(statusBarStyle.Render(statusText))
	b.WriteString("\n")

	// Agent details (path of current cursor item)
	if m.cursor >= 0 && m.cursor < len(m.items) && !m.items[m.cursor].isGroup {
		item := m.items[m.cursor]
		if m.pathValid[item.index] {
			b.WriteString(dimStyle.Render(item.agent.Path))
		} else {
			b.WriteString(warningStyle.Render("路徑不存在: " + item.agent.Path))
		}
		b.WriteString("\n")
	}

	// Help line
	b.WriteString(m.renderHelpBar())
	b.WriteString("\n")

	return b.String()
}

// viewConfirm renders the launch confirmation screen.
func (m Model) viewConfirm() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GST Agent Launcher"))
	b.WriteString("\n\n")

	agents := m.selectedAgents()
	total := len(agents)
	if m.monitorOn {
		total++
	}

	b.WriteString(fmt.Sprintf("即將開啟 %d 個 tab：\n\n", total))

	if m.monitorOn && m.config != nil {
		if m.monitorLaunched {
			b.WriteString(fmt.Sprintf("  %s Monitor (%s) %s\n",
				dimStyle.Render("●"), m.config.Monitor.Command,
				dimStyle.Render("— 已在運行，將跳過")))
		} else {
			b.WriteString(fmt.Sprintf("  %s Monitor (%s)\n",
				selectedStyle.Render("●"), m.config.Monitor.Command))
		}
	}

	for _, agent := range agents {
		valid := true
		running := false
		for i, a := range m.config.Agents {
			if a.Name == agent.Name {
				if !m.pathValid[i] {
					valid = false
				}
				if m.runningAgents[i] {
					running = true
				}
				break
			}
		}
		if !valid {
			b.WriteString(fmt.Sprintf("  %s %s [%s] %s\n",
				dimStyle.Render("●"), invalidStyle.Render(agent.Name), agent.Group,
				warningStyle.Render("(路徑不存在，將跳過)")))
		} else if running {
			b.WriteString(fmt.Sprintf("  %s %s [%s] %s\n",
				selectedStyle.Render("●"), agent.Name, agent.Group,
				warningStyle.Render("(已在運行，將重新開啟)")))
		} else {
			b.WriteString(fmt.Sprintf("  %s %s [%s]\n",
				selectedStyle.Render("●"), agent.Name, agent.Group))
		}
	}

	b.WriteString("\n")
	b.WriteString(m.renderHelpBar())
	b.WriteString("\n")

	return b.String()
}

// viewResult renders the launch result screen.
func (m Model) viewResult() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GST Agent Launcher"))
	b.WriteString("\n\n")

	if m.result == nil {
		b.WriteString("Launching...\n")
		return b.String()
	}

	if len(m.result.Launched) > 0 {
		b.WriteString(successStyle.Render(fmt.Sprintf("✓ 已啟動 %d 個 tab", len(m.result.Launched))))
		b.WriteString("\n\n")
		for _, name := range m.result.Launched {
			b.WriteString(fmt.Sprintf("  %s %s\n", selectedStyle.Render("●"), name))
		}
	}

	if len(m.result.Warnings) > 0 {
		b.WriteString("\n")
		b.WriteString(warningStyle.Render("⚠ 警告："))
		b.WriteString("\n")
		for _, w := range m.result.Warnings {
			b.WriteString(fmt.Sprintf("  %s\n", w))
		}
	}

	b.WriteString("\n")
	b.WriteString(m.renderHelpBar())
	b.WriteString("\n")

	return b.String()
}

// viewProjectSelect renders the project selection screen.
func (m Model) viewProjectSelect() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GST Agent Launcher — 專案選取"))
	b.WriteString("\n\n")

	for i, name := range m.projectNames {
		cursor := "  "
		if i == m.projectCursor {
			cursor = cursorStyle.Render("> ")
		}

		proj := m.config.Projects[name]
		label := name
		if i == m.projectCursor {
			label = cursorStyle.Render(name)
		}
		desc := dimStyle.Render(fmt.Sprintf("(%d agents) %s", len(proj.Agents), proj.Description))

		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, desc))
	}

	b.WriteString("\n")
	b.WriteString(m.renderHelpBar())
	b.WriteString("\n")

	return b.String()
}

// viewHelpOverlay renders the help overlay screen.
func (m Model) viewHelpOverlay() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GST Agent Launcher — 快捷鍵"))
	b.WriteString("\n\n")

	b.WriteString(confirmStyle.Render("導航"))
	b.WriteString("\n")
	b.WriteString("  ↑/k       上移\n")
	b.WriteString("  ↓/j       下移\n")
	b.WriteString("\n")

	b.WriteString(confirmStyle.Render("選取"))
	b.WriteString("\n")
	b.WriteString("  Space     勾選/取消\n")
	b.WriteString("  a         全選/全取消\n")
	b.WriteString("  Esc       清除所有選擇\n")
	b.WriteString("  r         恢復上次選擇\n")
	b.WriteString("\n")

	b.WriteString(confirmStyle.Render("群組快選"))
	b.WriteString("\n")
	b.WriteString("  c         Core 群組\n")
	b.WriteString("  p         PM 群組\n")
	b.WriteString("  o         App 群組\n")
	b.WriteString("  l         Leyu 群組\n")
	b.WriteString("\n")

	b.WriteString(confirmStyle.Render("功能"))
	b.WriteString("\n")
	b.WriteString("  Enter     啟動選取的 Agent\n")
	b.WriteString("  m         切換 Monitor\n")
	b.WriteString("  M         單獨啟動 Monitor\n")
	b.WriteString("  /         搜尋過濾\n")
	b.WriteString("  P         專案快選\n")
	b.WriteString("  e         編輯設定檔\n")
	b.WriteString("  ?         顯示此幫助\n")
	b.WriteString("  q         退出\n")
	b.WriteString("\n")

	b.WriteString(m.renderHelpBar())
	b.WriteString("\n")

	return b.String()
}

// renderHelpBar returns the help bar text for the current view state.
func (m Model) renderHelpBar() string {
	switch m.view {
	case viewConfirm:
		return helpStyle.Render("y:確認 n:取消")
	case viewResult:
		return helpStyle.Render("任意鍵:返回選單 q:退出")
	case viewHelp:
		return helpStyle.Render("按任意鍵關閉")
	case viewProject:
		return helpStyle.Render("↑↓:選擇 Enter:確認 Esc:取消")
	default:
		if m.searchMode {
			return helpStyle.Render("輸入搜尋 | Esc:清除 Enter:確認 Space:勾選")
		}
		return helpStyle.Render("↑↓/jk:導航 Space:勾選 Enter:啟動 /:搜尋 ?:幫助 M:Monitor Esc:清除 q:退出")
	}
}
