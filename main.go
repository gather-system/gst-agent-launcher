package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/gather-system/gst-agent-launcher/config"
	"github.com/gather-system/gst-agent-launcher/tui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("gst-launcher v%s (%s, %s)\n", version, commit, date)
		return
	}

	// Ensure user config directory exists on first run.
	config.EnsureUserConfig()

	m := tui.NewModel()
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error running program:", err)
		os.Exit(1)
	}
}
