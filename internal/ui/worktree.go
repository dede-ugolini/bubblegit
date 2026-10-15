package ui

import (
	"strings"

	"bubblegit/internal/git"

	"charm.land/lipgloss/v2"
)

// renderWorktrees draws the worktrees subpanel, which shares the files
// panel's slot and is toggled with the files tree by "w". It is
// display-only: there is no cursor action beyond moving the selection, so
// rows are plain text (the highlighted one swapping colors for the cursor
// background, as in the other list panels).
func (m *Model) renderWorktrees() string {
	p := m.panels[focusStag]
	if p.height <= 0 || p.width <= 0 {
		return ""
	}

	// The row is only the cursor when the files panel actually holds focus;
	// otherwise the subpanel is shown but not active, same as the other
	// panels' selection styling.
	selected := m.focus == focusStag

	start, end := visibleWindow(len(m.worktrees), p.height, p.idx)
	hashColor := lipgloss.NewStyle().Foreground(m.theme.Hash)
	branchColor := lipgloss.NewStyle().Foreground(m.theme.Accent)
	branch := func(w git.WorktreeEntry) string {
		name := w.Branch
		if name == "" {
			name = "(detached)"
		}
		return name
	}

	var rows []string
	for i := start; i < end; i++ {
		w := m.worktrees[i]
		if i == p.idx && selected {
			// Deliberately plain text here: coloring the fields first and
			// only then wrapping in Width(...).Background(...) would hit the
			// nested-reset issue seen in log/stash - see there.
			line := w.ShortHash + " " + branch(w) + " " + w.Path
			rows = append(rows, lipgloss.NewStyle().Width(p.width).Background(m.theme.Cursor).Render(line))
			continue
		}
		line := hashColor.Render(w.ShortHash) + " " + branchColor.Render(branch(w)) + " " + w.Path
		rows = append(rows, lipgloss.NewStyle().Width(p.width).Render(line))
	}
	// Same empty-list placeholder as the stash panel: the border needs one
	// padded line to hold its width.
	if len(rows) == 0 {
		rows = append(rows, lipgloss.NewStyle().Width(p.width).Render(""))
	}

	return m.titledPanel(focusStag, "Worktrees", 0, p.height, strings.Join(rows, "\n"))
}
