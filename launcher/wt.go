package launcher

import (
	"fmt"
	"os/exec"

	"github.com/gather-system/gst-agent-launcher/config"
)

// LaunchInWindowsTerminal opens a new Windows Terminal tab for the given agent.
func LaunchInWindowsTerminal(agent config.Agent) error {
	cmd := exec.Command("wt", "-w", "0", "new-tab",
		"--title", agent.Name,
		"-d", agent.Path,
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch Windows Terminal for %s: %w", agent.Name, err)
	}

	return nil
}
