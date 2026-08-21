package ui

import (
	"strings"

	"bubblegit/internal/git"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	if m.quitting {
		tea.NewView("")
	}

	// TODO: add spinner
	if !m.ready {
		tea.NewView("Loading...")
	}

	return tea.NewView(renderBranches(m.dir, m.branchFocus, m.focus))
}

func renderBranches(dir string, idx, focus int) string {
	branches, _ := git.Branches(dir)
	var names []string

	for i, b := range branches {
		name := b.Name
		if idx == i {
			name = lipgloss.NewStyle().Background(lipgloss.ANSIColor(9)).Render(name)
		}
		if b.Current {
			name = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(11)).Render(name)
		}
		names = append(names, name)
	}

	if focus == focusBranch {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(lipgloss.ANSIColor(222)).
			Render(strings.Join(names, "\n"))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Render(strings.Join(names, "\n"))
}
