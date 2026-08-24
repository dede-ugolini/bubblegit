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

	b.WriteString(renderFiles(m.files, m.idxFiles, m.focus))
	b.WriteString("\n")

	b.WriteString(renderBranches(m.branches, m.branchFocus, m.focus))
	b.WriteString("\n")

	if m.focusInput {
		b.WriteString(m.input.View())
		b.WriteString("\n")
	}
	b.WriteString(renderFooter(m.focus))

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(m.err.Error()))
	}
	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func renderFiles(files []git.FileStatus, idx, focus int) string {
	var s []string
	red := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(lipgloss.Red))
	green := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(lipgloss.Green))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(lipgloss.Yellow))
	sel := lipgloss.NewStyle().Background(lipgloss.ANSIColor(11))

	for i, f := range files {
		stag := string(f.Index)
		worktree := string(f.Worktree)
		path := f.Path

		switch {
		case f.Untracked():
			stag = red.Render(stag)
			worktree = red.Render(worktree)

		case f.Staged():
			stag = green.Render(stag)
			if f.Worktree == 'M' {
				worktree = red.Render(worktree)
				path = yellow.Render(path)
			} else {
				path = green.Render(path)
			}

		case f.Unstaged():
			worktree = red.Render(worktree)
		}

		space := " "

		if i == idx && focus == focusStag {
			stag = sel.Render(stag)
			worktree = sel.Render(worktree)
			space = sel.Render(space)
			path = sel.Render(path)
		}

		s = append(s, stag+worktree+space+path)
	}

	if focus == focusStag {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(lipgloss.ANSIColor(222)).
			Render(strings.Join(s, "\n"))
	}

	style := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true)
	return style.Render(strings.Join(s, "\n"))
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
	case focusStag:
		return helpStyle.Render("↑/k ↓/j move · <space> stag · q quit")
	case focusBranch:
		return helpStyle.Render("↑/k ↓/j move · space stage · enter checkout · n new branch · r rename · d delete · q quit")
	default:
		return ""
	}
}
