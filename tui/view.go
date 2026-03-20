package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// View renders the TUI.
func (m Model) View() tea.View {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GST Agent Launcher"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("q: quit"))
		b.WriteString("\n")
		return tea.NewView(b.String())
	}

	if m.config == nil {
		b.WriteString("Loading...\n")
		return tea.NewView(b.String())
	}

	if len(m.items) == 0 {
		b.WriteString("No agents configured.\n")
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("q: quit"))
		b.WriteString("\n")
		return tea.NewView(b.String())
	}

	// Render list items
	for i, item := range m.items {
		if item.isGroup {
			// Group header
			style := groupStyle(item.group)
			b.WriteString(style.Render(fmt.Sprintf("── %s ──", item.group)))
			b.WriteString("\n")
			continue
		}

		// Cursor
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		// Checkbox
		check := "[ ]"
		if m.selected[item.index] {
			check = selectedStyle.Render("[x]")
		}

		// Agent name
		name := item.agent.Name
		if i == m.cursor {
			name = cursorStyle.Render(name)
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, name))
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

	return tea.NewView(b.String())
}
