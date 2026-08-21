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

	return tea.NewView(renderBranches(m.dir))
}

func renderBranches(dir string) string {
	branches, _ := git.Branches(dir)
	var names []string
	for _, b := range branches {
		if b.Current {
			names = append(names, lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(11)).Render(b.Name))
			continue
		}
		names = append(names, b.Name)
		// TODO: style current branch
	}
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).Render(strings.Join(names, "\n"))
}
