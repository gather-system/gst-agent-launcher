package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/gather-system/gst-agent-launcher/config"
)

// groupOrder defines the display order of groups.
var groupOrder = []string{"PM", "Core", "App", "Leyu"}

// listItem represents a single row in the TUI list.
type listItem struct {
	isGroup bool          // true = group header (not selectable)
	group   string        // group name
	agent   *config.Agent // non-nil for agent items
	index   int           // index into config.Agents
}

// Model is the Bubble Tea model for the launcher TUI.
type Model struct {
	config    *config.Config
	items     []listItem    // flat list of group headers + agents
	cursor    int           // current cursor position in items
	selected  map[int]bool  // keyed by agent index in config.Agents
	monitorOn bool          // monitor toggle
	err       error
}

// NewModel creates a new Model with default state.
func NewModel() Model {
	return Model{
		selected: make(map[int]bool),
	}
}

// Init loads the configuration on startup.
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			return errMsg{err}
		}
		return configLoadedMsg{cfg}
	}
}

// buildItems creates the flat list of group headers and agent items.
func buildItems(cfg *config.Config) []listItem {
	// Group agents by group name.
	grouped := make(map[string][]int)
	for i, agent := range cfg.Agents {
		grouped[agent.Group] = append(grouped[agent.Group], i)
	}

	var items []listItem
	for _, group := range groupOrder {
		indices, ok := grouped[group]
		if !ok || len(indices) == 0 {
			continue
		}
		items = append(items, listItem{isGroup: true, group: group})
		for _, idx := range indices {
			items = append(items, listItem{
				isGroup: false,
				group:   group,
				agent:   &cfg.Agents[idx],
				index:   idx,
			})
		}
	}
	return items
}

// firstSelectableIndex returns the index of the first selectable item.
func firstSelectableIndex(items []listItem) int {
	for i, item := range items {
		if !item.isGroup {
			return i
		}
	}
	return 0
}

// selectedCount returns the number of selected agents.
func (m Model) selectedCount() int {
	count := 0
	for _, v := range m.selected {
		if v {
			count++
		}
	}
	return count
}

// configLoadedMsg is sent when the config has been loaded.
type configLoadedMsg struct {
	config *config.Config
}

// errMsg is sent when an error occurs.
type errMsg struct {
	err error
}
