package tui

import (
	"fmt"
	"strings"
	"time"

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
	case viewDashboard:
		content = m.viewDashboard()
	case viewDeps:
		content = m.viewDepsPrompt()
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
		badge := m.healthBadge(item.index)
		gitInfo := m.gitStatusLabel(item.index)
		if !m.pathValid[item.index] {
			name = invalidStyle.Render(name + " [!]")
		} else if i == m.cursor {
			name = cursorStyle.Render(name)
		}
		if m.runningAgents[item.index] {
			name += " " + successStyle.Render("[R]")
		}
		if badge != "" {
			name += " " + badge
		}
		if gs, ok := m.gitStatuses[item.index]; ok && gs.HasOpenPR {
			name += " " + successStyle.Render("[PR]")
		}
		if gitInfo != "" {
			name += " " + gitInfo
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

	// Agent details (path + health info of current cursor item)
	if m.cursor >= 0 && m.cursor < len(m.items) && !m.items[m.cursor].isGroup {
		item := m.items[m.cursor]
		if !m.pathValid[item.index] {
			b.WriteString(warningStyle.Render("路徑不存在: " + item.agent.Path))
		} else {
			b.WriteString(dimStyle.Render(item.agent.Path))
			if gs, ok := m.gitStatuses[item.index]; ok && gs.IssueID != "" {
				b.WriteString("  " + confirmStyle.Render(gs.IssueID))
			}
			if hr, ok := m.healthResults[item.index]; ok {
				if !hr.IsGitRepo {
					b.WriteString("  " + dimStyle.Render("非 Git 倉庫"))
				} else if hr.HasConflict {
					b.WriteString("  " + warningStyle.Render("有未解決的 merge conflict"))
				}
			}
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

// viewDashboard renders the dashboard table view.
func (m Model) viewDashboard() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GST Agent Dashboard"))
	b.WriteString("\n\n")

	if m.config == nil {
		b.WriteString("Loading...\n")
		return b.String()
	}

	// Table header.
	header := fmt.Sprintf("  %-20s %-6s %-8s %-30s %-6s %-4s %-8s",
		"Name", "Group", "Status", "Branch", "Dirty", "PR", "Health")
	b.WriteString(dashHeaderStyle.Render(header))
	b.WriteString("\n")

	// Table rows.
	for i, agent := range m.config.Agents {
		status := dimStyle.Render("Stopped")
		if m.runningAgents[i] {
			status = successStyle.Render("Running")
		}

		branch := ""
		dirty := ""
		if gs, ok := m.gitStatuses[i]; ok {
			branch = gs.Branch
			if len(branch) > 28 {
				branch = branch[:28] + ".."
			}
			if gs.DirtyCount > 0 {
				dirty = fmt.Sprintf("*%d", gs.DirtyCount)
			}
		} else if m.gitLoading && m.pathValid[i] {
			branch = "..."
		}

		healthStr := ""
		if !m.pathValid[i] {
			healthStr = "[!]"
		} else if hr, ok := m.healthResults[i]; ok {
			if !hr.IsGitRepo {
				healthStr = "[!git]"
			} else if hr.HasConflict {
				healthStr = "[warn]"
			}
		}

		prStr := ""
		if gs, ok := m.gitStatuses[i]; ok && gs.HasOpenPR {
			prStr = successStyle.Render("✓")
		}

		style := groupStyle(agent.Group)
		row := fmt.Sprintf("  %-20s %-6s %-8s %-30s %-6s %-4s %-8s",
			agent.Name, style.Render(agent.Group), status, branch, dirty, prStr, healthStr)

		if i%2 == 0 {
			b.WriteString(row)
		} else {
			b.WriteString(dashRowAltStyle.Render(row))
		}
		b.WriteString("\n")
	}

	// Footer.
	b.WriteString("\n")
	now := time.Now().Format("15:04:05")
	b.WriteString(dimStyle.Render(fmt.Sprintf("上次刷新: %s | 每 30 秒自動刷新", now)))
	b.WriteString("\n")
	b.WriteString(m.renderHelpBar())
	b.WriteString("\n")

	return b.String()
}

// viewDepsPrompt renders the dependency prompt screen.
func (m Model) viewDepsPrompt() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GST Agent Launcher"))
	b.WriteString("\n\n")

	b.WriteString(warningStyle.Render("以下依賴尚未選取或運行："))
	b.WriteString("\n\n")

	for _, dep := range m.unmetDeps {
		b.WriteString(fmt.Sprintf("  %s %s\n", warningStyle.Render("●"), dep))
	}

	b.WriteString("\n")
	b.WriteString("是否一併啟動？\n")
	b.WriteString("\n")
	b.WriteString(m.renderHelpBar())
	b.WriteString("\n")

	return b.String()
}

// renderHelpBar returns the help bar text for the current view state.
func (m Model) renderHelpBar() string {
	switch m.view {
	case viewDashboard:
		return helpStyle.Render("d:返回清單 q:退出")
	case viewDeps:
		return helpStyle.Render("y:一併啟動 n:跳過 Esc:返回")
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
		return helpStyle.Render("↑↓/jk:導航 Space:勾選 Enter:啟動 d:Dashboard /:搜尋 ?:幫助 q:退出")
	}
}
