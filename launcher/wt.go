package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gather-system/gst-agent-launcher/config"
)

// LaunchResult holds the outcome of a batch launch.
type LaunchResult struct {
	Launched []string // names of successfully queued agents
	Warnings []string // warnings (e.g. path not found)
}

// LaunchAll opens all selected agents (and optionally the monitor) in a single
// wt.exe invocation using semicolon-separated new-tab commands.
func LaunchAll(agents []config.Agent, monitor *config.Monitor) (*LaunchResult, error) {
	wtPath, err := findWT()
	if err != nil {
		return nil, err
	}

	result := &LaunchResult{}
	var argSets [][]string

	// Monitor tab first (if enabled and monitor is provided).
	if monitor != nil && monitor.Enabled && monitor.Command != "" {
		parts := strings.Fields(monitor.Command)
		monitorArgs := []string{"new-tab", "--title", "Monitor"}
		monitorArgs = append(monitorArgs, "pwsh", "-NoExit", "-Command")
		monitorArgs = append(monitorArgs, parts...)
		argSets = append(argSets, monitorArgs)
		result.Launched = append(result.Launched, "Monitor")
	}

	// Agent tabs.
	for _, agent := range agents {
		if _, err := os.Stat(agent.Path); os.IsNotExist(err) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("%s: path not found (%s)", agent.Name, agent.Path))
			continue
		}
		tabArgs := []string{"new-tab", "-d", agent.Path, "--title", agent.Name}
		argSets = append(argSets, tabArgs)
		result.Launched = append(result.Launched, agent.Name)
	}

	if len(argSets) == 0 {
		return result, nil
	}

	// Build the full argument list: wt -w 0 <first-tab> ; <second-tab> ; ...
	var args []string
	args = append(args, "-w", "0")
	for i, set := range argSets {
		if i > 0 {
			args = append(args, ";")
		}
		args = append(args, set...)
	}

	cmd := exec.Command(wtPath, args...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start wt.exe: %w", err)
	}

	// Release the process handle in the background.
	go cmd.Wait()

	return result, nil
}

// LaunchMonitor opens only the monitor tab in Windows Terminal.
func LaunchMonitor(monitor config.Monitor) error {
	wtPath, err := findWT()
	if err != nil {
		return err
	}

	parts := strings.Fields(monitor.Command)
	args := []string{"-w", "0", "new-tab", "--title", "Monitor", "pwsh", "-NoExit", "-Command"}
	args = append(args, parts...)

	cmd := exec.Command(wtPath, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start wt.exe: %w", err)
	}
	go cmd.Wait()
	return nil
}

// findWT locates wt.exe, checking the standard WindowsApps path first,
// then falling back to PATH lookup.
func findWT() (string, error) {
	// Try standard location.
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		candidate := filepath.Join(localAppData, "Microsoft", "WindowsApps", "wt.exe")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// Fall back to PATH.
	path, err := exec.LookPath("wt.exe")
	if err != nil {
		return "", fmt.Errorf("wt.exe not found — please install Windows Terminal from the Microsoft Store")
	}
	return path, nil
}
