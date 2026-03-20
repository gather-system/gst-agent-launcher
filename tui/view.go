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
	} else if m.config == nil {
		b.WriteString("Loading...\n")
	} else if len(m.config.Agents) == 0 {
		b.WriteString("No agents configured.\n")
	} else {
		for _, agent := range m.config.Agents {
			b.WriteString(fmt.Sprintf("  %s (%s)\n", agent.Name, agent.Group))
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("q: quit"))
	b.WriteString("\n")

	return tea.NewView(b.String())
}
