package tui

import tea "charm.land/bubbletea/v2"

// Update handles incoming messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case configLoadedMsg:
		m.config = msg.config
		m.items = buildItems(m.config)
		m.cursor = firstSelectableIndex(m.items)
		m.monitorOn = m.config.Monitor.Enabled
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			m.moveCursorUp()

		case "down", "j":
			m.moveCursorDown()

		case "space", " ":
			m.toggleCurrent()

		case "enter":
			return m, tea.Quit

		case "a":
			m.toggleAll()

		case "c":
			m.toggleGroup("Core")

		case "p":
			m.toggleGroup("PM")

		case "o":
			m.toggleGroup("App")

		case "l":
			m.toggleGroup("Leyu")

		case "m":
			m.monitorOn = !m.monitorOn
		}
	}

	return m, nil
}

// moveCursorUp moves the cursor up, skipping group headers.
func (m *Model) moveCursorUp() {
	for i := m.cursor - 1; i >= 0; i-- {
		if !m.items[i].isGroup {
			m.cursor = i
			return
		}
	}
}

// moveCursorDown moves the cursor down, skipping group headers.
func (m *Model) moveCursorDown() {
	for i := m.cursor + 1; i < len(m.items); i++ {
		if !m.items[i].isGroup {
			m.cursor = i
			return
		}
	}
}

// toggleCurrent toggles the selection of the item at the cursor.
func (m *Model) toggleCurrent() {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return
	}
	item := m.items[m.cursor]
	if item.isGroup {
		return
	}
	m.selected[item.index] = !m.selected[item.index]
}

// toggleAll selects all agents if any are unselected, otherwise deselects all.
func (m *Model) toggleAll() {
	if m.config == nil {
		return
	}
	allSelected := true
	for i := range m.config.Agents {
		if !m.selected[i] {
			allSelected = false
			break
		}
	}
	for i := range m.config.Agents {
		m.selected[i] = !allSelected
	}
}

// toggleGroup selects all agents in a group if any are unselected, otherwise deselects all.
func (m *Model) toggleGroup(group string) {
	if m.config == nil {
		return
	}
	var indices []int
	for i, agent := range m.config.Agents {
		if agent.Group == group {
			indices = append(indices, i)
		}
	}
	allSelected := true
	for _, i := range indices {
		if !m.selected[i] {
			allSelected = false
			break
		}
	}
	for _, i := range indices {
		m.selected[i] = !allSelected
	}
}
