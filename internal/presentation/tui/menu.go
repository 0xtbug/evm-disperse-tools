package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// menuEntry represents a single menu item
type menuEntry struct {
	id    int
	label string
	desc  string
}

// Menu entries for the disperse tool
var menuEntries = []menuEntry{
	{1, "Disperse Native", "Send native tokens to multiple recipients"},
	{2, "Disperse ERC20", "Send ERC20 tokens to multiple recipients"},
	{3, "View Reports", "Check past execution reports and history"},
	{4, "Fee Calculator", "Estimate gas costs for bulk native transfers"},
}

// MenuModel is the main menu screen
type MenuModel struct {
	cursor int
}

// NewMenuModel creates a new menu model
func NewMenuModel() MenuModel {
	return MenuModel{cursor: 0}
}

// Update handles keyboard and mouse input
func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(menuEntries)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m, m.dispatch(menuEntries[m.cursor].id)
		case "1", "2", "3", "4":
			// Direct number selection
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(menuEntries) {
				m.cursor = idx
				return m, m.dispatch(menuEntries[idx].id)
			}
		}
	}
	return m, nil
}

// HandleMouseClick processes a mouse click at the given Y position
// Returns true if a menu item was clicked, along with the command
func (m *MenuModel) HandleMouseClick(mouseY int, menuStartY int) (bool, tea.Cmd) {
	itemIndex := mouseY - menuStartY
	if itemIndex >= 0 && itemIndex < len(menuEntries) {
		m.cursor = itemIndex
		return true, m.dispatch(menuEntries[itemIndex].id)
	}
	return false, nil
}

// View renders the menu
func (m MenuModel) View(width int) string {
	var sb strings.Builder

	for i := 0; i < len(menuEntries); i++ {
		row := m.entryView(i, width)
		sb.WriteString(row)
	}

	return strings.TrimRight(sb.String(), "\n")
}

// entryView renders a single menu entry
func (m MenuModel) entryView(i int, width int) string {
	e := menuEntries[i]

	// Pad label to consistent width
	paddedLabel := fmt.Sprintf("%-20s", e.label)

	var row string
	if m.cursor == i {
		// Selected: show with > prefix and selection background
		row = "  " + SelectedStyle.Render(fmt.Sprintf("> %s  %s", paddedLabel, e.desc))
	} else {
		// Unselected: normal styling
		row = fmt.Sprintf("    %s  %s",
			ToolLabelStyle.Render(paddedLabel),
			ToolDescStyle.Render(e.desc))
	}

	return row + "\n"
}

// dispatch sends a runToolMsg for the selected menu item
func (m MenuModel) dispatch(id int) tea.Cmd {
	return func() tea.Msg { return runToolMsg{toolID: id} }
}

// MenuEntries returns the menu entries (for tests)
func MenuEntries() []string {
	entries := make([]string, len(menuEntries))
	for i, e := range menuEntries {
		entries[i] = e.label
	}
	return entries
}

// SelectedIsExit returns true if the current selection is Exit
func (m MenuModel) SelectedIsExit() bool {
	return m.cursor == len(menuEntries)-1
}

// GetSelectedID returns the ID of the currently selected entry
func (m MenuModel) GetSelectedID() int {
	return menuEntries[m.cursor].id
}

// GetCursor returns the current cursor position
func (m MenuModel) GetCursor() int {
	return m.cursor
}

// SetCursor sets the cursor position
func (m *MenuModel) SetCursor(cursor int) {
	if cursor >= 0 && cursor < len(menuEntries) {
		m.cursor = cursor
	}
}
