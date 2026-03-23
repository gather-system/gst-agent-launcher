package tui

import "charm.land/lipgloss/v2"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2)

	// Group header styles
	groupStylePM = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF00FF"))

	groupStyleCore = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00"))

	groupStyleApp = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00AAFF"))

	groupStyleLeyu = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFF00"))

	// Item styles
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF00FF"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	confirmStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFF00"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00"))

	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6600"))

	toastStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1A1A1A")).
			Background(lipgloss.Color("#FFCC00")).
			Padding(0, 1)
)

// groupStyle returns the lipgloss style for a given group name.
func groupStyle(group string) lipgloss.Style {
	switch group {
	case "PM":
		return groupStylePM
	case "Core":
		return groupStyleCore
	case "App":
		return groupStyleApp
	case "Leyu":
		return groupStyleLeyu
	default:
		return normalStyle
	}
}
