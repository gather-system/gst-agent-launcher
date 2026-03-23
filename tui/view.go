package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// View renders the TUI.
func (m Model) View() tea.View {
	switch m.view {
	case viewConfirm:
		return tea.NewView(m.viewConfirm())
	case viewResult:
		return tea.NewView(m.viewResult())
	default:
		return tea.NewView(m.viewList())
	}
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

	// Render list items
	for i, item := range m.items {
		if item.isGroup {
			style := groupStyle(item.group)
			b.WriteString(style.Render(fmt.Sprintf("── %s ──", item.group)))
			b.WriteString("\n")
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
		if i == m.cursor {
			name = cursorStyle.Render(name)
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
	monitorStatus := "OFF"
	if m.monitorOn {
		monitorStatus = "ON"
	}
	statusText := fmt.Sprintf("已選: %d 個 Agent | Monitor: %s | Enter=啟動 q=退出",
		m.selectedCount(), monitorStatus)
	b.WriteString(statusBarStyle.Render(statusText))
	b.WriteString("\n")

	// Help line
	b.WriteString(helpStyle.Render("↑↓/jk:導航 Space:勾選 a:全選 c:Core p:PM o:App l:Leyu m:Monitor"))
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
		b.WriteString(fmt.Sprintf("  %s %s [%s]\n",
			selectedStyle.Render("●"), agent.Name, agent.Group))
	}

	b.WriteString("\n")
	b.WriteString(confirmStyle.Render("按 y 確認啟動 / n 返回選單"))
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
	b.WriteString(helpStyle.Render("按任意鍵返回選單 / q 退出"))
	b.WriteString("\n")

	return b.String()
}
