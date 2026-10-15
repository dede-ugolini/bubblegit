package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) renderTags() string {
	p := m.panels[focusBranch]
	if p.height <= 0 || p.width <= 0 {
		return ""
	}

	// The row is only the cursor when the branch panel actually holds focus;
	// otherwise the subpanel is shown but not active, same as the other
	// panels' selection styling.
	selected := m.focus == focusBranch

	start, end := visibleWindow(len(m.tagList), p.height, p.idx)
	nameColor := lipgloss.NewStyle().Foreground(m.theme.Accent)

	var rows []string
	for i := start; i < end; i++ {
		t := m.tagList[i]
		if i == p.idx && selected {
			// Deliberately plain text here: coloring the name first and only
			// then wrapping in Width(...).Background(...) would hit the
			// nested-reset issue seen in log/stash - see there.
			rows = append(rows, lipgloss.NewStyle().Width(p.width).Background(m.theme.Cursor).Render(t.Name+" "+t.Message))
			continue
		}
		rows = append(rows, lipgloss.NewStyle().Width(p.width).Render(nameColor.Render(t.Name)+" "+t.Message))
	}
	// Same empty-list placeholder as the worktree panel: the border needs one
	// padded line to hold its width.
	if len(rows) == 0 {
		rows = append(rows, lipgloss.NewStyle().Width(p.width).Render(""))
	}

	return m.titledPanel(focusBranch, "Tags", 0, p.height, strings.Join(rows, "\n"))
}
