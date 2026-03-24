package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/gather-system/gst-agent-launcher/config"
	gitpkg "github.com/gather-system/gst-agent-launcher/git"
	"github.com/gather-system/gst-agent-launcher/health"
	"github.com/gather-system/gst-agent-launcher/launcher"
)

// viewState represents the current TUI screen.
type viewState int

const (
	viewList      viewState = iota // agent selection list
	viewConfirm                   // launch confirmation
	viewResult                    // launch result
	viewHelp                      // help overlay
	viewProject                   // project selection
	viewDashboard                 // dashboard table view
	viewDeps                      // dependency prompt
	viewGitResult                 // batch git operation result
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
	config          *config.Config
	items           []listItem    // flat list of group headers + agents
	cursor          int           // current cursor position in items
	selected        map[int]bool  // keyed by agent index in config.Agents
	pathValid       map[int]bool  // keyed by agent index, true if path exists
	healthResults   map[int]health.CheckResult // keyed by agent index
	gitAvailable    bool          // true if git is installed
	gitStatuses     map[int]gitpkg.RepoStatus // keyed by agent index
	gitLoading      bool          // true while git status is being fetched
	runningAgents   map[int]bool  // keyed by agent index, true if process detected
	dashboardTimer  int           // dashboard tick generation ID
	batchResults    []gitpkg.BatchResult // results from batch git operation
	batchLoading    bool          // true while batch operation is running
	monitorOn       bool          // monitor toggle
	monitorLaunched bool          // true after Monitor has been launched
	view            viewState     // current screen
	result          *launcher.LaunchResult
	err             error
	toast           string // current toast message (empty = no toast)
	toastTimer      int    // toast generation ID for cancelling stale timers
	searchMode      bool     // true when in search/filter mode
	searchQuery     string   // current search query
	projectNames    []string // sorted project names for display
	projectCursor   int      // cursor position in project list
	lastClickY      int      // last mouse click Y for double-click detection
	lastClickTime   int64    // last mouse click time (UnixMilli) for double-click
	unmetDeps       []string // unmet dependency agent names (for viewDeps prompt)
}

// NewModel creates a new Model with default state.
func NewModel() Model {
	return Model{
		selected: make(map[int]bool),
		view:     viewList,
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

// selectedAgents returns the list of selected agents in display order.
func (m Model) selectedAgents() []config.Agent {
	var agents []config.Agent
	for _, item := range m.items {
		if !item.isGroup && m.selected[item.index] {
			agents = append(agents, *item.agent)
		}
	}
	return agents
}

// buildItems creates the flat list of group headers and agent items.
func buildItems(cfg *config.Config) []listItem {
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


// resetSelection clears all agent selections, keeping monitorOn intact.
func (m *Model) resetSelection() {
	m.selected = make(map[int]bool)
}

// selectedAgentNames returns the names of all selected agents.
func (m Model) selectedAgentNames() []string {
	var names []string
	for _, item := range m.items {
		if !item.isGroup && m.selected[item.index] {
			names = append(names, item.agent.Name)
		}
	}
	return names
}

// restoreSession applies a saved session by matching agent names.
func (m *Model) restoreSession(session *config.Session) {
	m.selected = make(map[int]bool)
	if m.config == nil {
		return
	}
	nameSet := make(map[string]bool)
	for _, name := range session.Agents {
		nameSet[name] = true
	}
	for i, agent := range m.config.Agents {
		if nameSet[agent.Name] {
			m.selected[i] = true
		}
	}
	m.monitorOn = session.MonitorOn
}

// itemAtY returns the items index for a given screen Y coordinate in viewList, or -1.
func (m Model) itemAtY(y int) int {
	row := 2 // title + blank line
	if m.monitorLaunched {
		row += 2 // Monitor [R] + blank line
	}
	for i, item := range m.items {
		if item.isGroup {
			if m.searchQuery != "" {
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
			}
		} else if !m.matchesSearch(item) {
			continue
		}
		if row == y {
			return i
		}
		row++
	}
	return -1
}

// buildProjectNames extracts sorted project names from config.
func buildProjectNames(cfg *config.Config) []string {
	if cfg == nil || len(cfg.Projects) == 0 {
		return nil
	}
	names := make([]string, 0, len(cfg.Projects))
	for name := range cfg.Projects {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// selectProject selects agents matching the given project and returns the count.
func (m *Model) selectProject(name string) int {
	if m.config == nil {
		return 0
	}
	proj, ok := m.config.Projects[name]
	if !ok {
		return 0
	}
	m.selected = make(map[int]bool)
	nameSet := make(map[string]bool)
	for _, a := range proj.Agents {
		nameSet[a] = true
	}
	count := 0
	for i, agent := range m.config.Agents {
		if nameSet[agent.Name] {
			m.selected[i] = true
			count++
		}
	}
	return count
}

// fuzzyMatch returns true if every character in query appears in target in order (case-insensitive).
func fuzzyMatch(query, target string) bool {
	query = strings.ToLower(query)
	target = strings.ToLower(target)
	qRunes := []rune(query)
	qi := 0
	for _, c := range target {
		if qi < len(qRunes) && c == qRunes[qi] {
			qi++
		}
	}
	return qi == len(qRunes)
}

// matchesSearch returns true if the item should be visible given the current search query.
func (m Model) matchesSearch(item listItem) bool {
	if m.searchQuery == "" {
		return true
	}
	if item.isGroup {
		return true // groups handled separately in view
	}
	return fuzzyMatch(m.searchQuery, item.agent.Name)
}

// adjustCursorForSearch moves the cursor to the first matching item if the current one is hidden.
func (m *Model) adjustCursorForSearch() {
	if m.cursor >= 0 && m.cursor < len(m.items) && !m.items[m.cursor].isGroup && m.matchesSearch(m.items[m.cursor]) {
		return
	}
	for i, item := range m.items {
		if !item.isGroup && m.matchesSearch(item) {
			m.cursor = i
			return
		}
	}
}

// healthBadge returns the health status badge for an agent, or "" if healthy.
func (m Model) healthBadge(agentIndex int) string {
	hr, ok := m.healthResults[agentIndex]
	if !ok {
		return ""
	}
	if !hr.PathValid {
		return "" // [!] is already shown by the existing invalidStyle logic
	}
	if !hr.IsGitRepo {
		return notGitStyle.Render("[!git]")
	}
	if hr.HasConflict {
		return conflictStyle.Render("[⚠]")
	}
	return ""
}

// gitStatusLabel returns a formatted git status string for an agent, or "" if not available.
func (m Model) gitStatusLabel(agentIndex int) string {
	if m.gitLoading && m.pathValid[agentIndex] {
		return dimStyle.Render("...")
	}
	gs, ok := m.gitStatuses[agentIndex]
	if !ok {
		return ""
	}
	label := dimStyle.Render(gs.Branch)
	if gs.DirtyCount > 0 {
		label += " " + warningStyle.Render(fmt.Sprintf("*%d", gs.DirtyCount))
	}
	return label
}

// checkDependencies returns unmet dependency agent names for all selected agents.
// A dependency is "met" if the dependency agent is already selected or running.
func (m Model) checkDependencies() []string {
	if m.config == nil {
		return nil
	}

	// Build name→index map.
	nameToIndex := make(map[string]int)
	for i, a := range m.config.Agents {
		nameToIndex[a.Name] = i
	}

	seen := make(map[string]bool)
	var unmet []string

	for i, agent := range m.config.Agents {
		if !m.selected[i] {
			continue
		}
		for _, dep := range agent.Dependencies {
			if seen[dep] {
				continue
			}
			idx, exists := nameToIndex[dep]
			if !exists {
				continue // unknown dependency, skip
			}
			if !m.selected[idx] && !m.runningAgents[idx] {
				unmet = append(unmet, dep)
				seen[dep] = true
			}
		}
	}
	return unmet
}

// selectDependencies selects all agents matching the given dependency names.
func (m *Model) selectDependencies(deps []string) {
	if m.config == nil {
		return
	}
	nameSet := make(map[string]bool)
	for _, d := range deps {
		nameSet[d] = true
	}
	for i, agent := range m.config.Agents {
		if nameSet[agent.Name] {
			m.selected[i] = true
		}
	}
}

// groupCount returns the number of selected and total agents in a group.
func (m Model) groupCount(group string) (selected, total int) {
	if m.config == nil {
		return 0, 0
	}
	for i, agent := range m.config.Agents {
		if agent.Group == group {
			total++
			if m.selected[i] {
				selected++
			}
		}
	}
	return
}
