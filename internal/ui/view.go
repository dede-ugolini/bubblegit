package ui

import (
	"strings"

	"bubblegit/internal/git"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var errStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(1))

func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	// TODO: add spinner
	if !m.ready {
		return tea.NewView("Loading...")
	}

	var b strings.Builder
	b.WriteString(renderBranches(m.branches, m.branchFocus, m.focus))
	b.WriteString("\n")
	b.WriteString(renderFooter(m.focus))

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(m.err.Error()))
	}
	return tea.NewView(b.String())
}

func renderBranches(branches []git.BranchInfo, idx, focus int) string {
	var names []string

	for i, b := range branches {
		name := b.Name
		if b.Current {
			name = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(11)).Render(name)
		}
		if idx == i {
			name = lipgloss.NewStyle().Background(lipgloss.ANSIColor(9)).Render(name)
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

func renderFooter(focus int) string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	switch focus {
	case focusBranch:
		return helpStyle.Render("↑/k ↓/j move · enter checkout · n new branch · d delete · q quit")
	}
	return helpStyle.Render("↑/k ↓/j move · enter checkout · n new branch · d delete · q quit")
}
