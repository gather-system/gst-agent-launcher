package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/gather-system/gst-agent-launcher/tui"
)

func main() {
	m := tui.NewModel()
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error running program:", err)
		os.Exit(1)
	}
}
