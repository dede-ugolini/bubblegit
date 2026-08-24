package ui

import (
	"strings"

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

	b.WriteString(m.renderFiles())
	b.WriteString("\n")

	b.WriteString(m.renderBranches())
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(m.renderLog())

	if m.focusInput {
		b.WriteString(m.input.View())
		b.WriteString("\n")
	}
	b.WriteString(renderFooter(m.focus))

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(m.err.Error()))
	}
	v := tea.NewView(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			b.String(), m.renderDiff(),
		))
	v.AltScreen = true
	return v
}

func (m *Model) renderDiff() string {
	if m.focus == focusDiff {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(lipgloss.ANSIColor(222)).
			Render(m.diff.View())
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Render(m.diff.View())
}

func (m *Model) renderFiles() string {
	var s []string
	red := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(lipgloss.Red))
	green := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(lipgloss.Green))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(lipgloss.Yellow))
	sel := lipgloss.NewStyle().Background(lipgloss.ANSIColor(11))

	for i, f := range m.files {
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

		if i == m.idxFiles && m.focus == focusStag {
			stag = sel.Render(stag)
			worktree = sel.Render(worktree)
			space = sel.Render(space)
			path = sel.Render(path)
		}

		s = append(s, stag+worktree+space+path)
	}

	if m.focus == focusStag {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(lipgloss.ANSIColor(222)).
			Width(m.filesWidth).
			Height(m.filesHeight).
			Render(strings.Join(s, "\n"))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Width(m.filesWidth).
		Height(m.filesHeight).
		Render(strings.Join(s, "\n"))
}

func (m *Model) renderBranches() string {
	var names []string

	for i, b := range m.branches {
		name := b.Name
		if b.Current {
			name = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(11)).Render(name)
		}
		if m.idxBranch == i {
			name = lipgloss.NewStyle().Background(lipgloss.ANSIColor(9)).Render(name)
		}
		names = append(names, name)
	}

	if m.focus == focusBranch {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(lipgloss.ANSIColor(222)).
			Width(m.branchWidth).
			Height(m.branchHeight).
			Render(strings.Join(names, "\n"))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Width(m.branchWidth).
		Height(m.branchHeight).
		Render(strings.Join(names, "\n"))
}

func (m *Model) renderLog() string {
	var entrys []string
	for i, l := range m.log {
		if i == m.idxLog && m.focus == focusLog {
			entrys = append(entrys, lipgloss.NewStyle().Background(lipgloss.ANSIColor(9)).Render(l.ShortHash))
			continue
		}
		entrys = append(entrys, l.ShortHash)
	}
	if m.focus == focusLog {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(lipgloss.ANSIColor(222)).
			Width(m.logWidth).
			Height(m.logHeight).
			Render(strings.Join(entrys, "\n"))
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Width(m.logWidth).
		Height(m.logHeight).
		Render(strings.Join(entrys, "\n"))
}

func renderFooter(focus int) string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	switch focus {
	case focusStag:
		return helpStyle.Render("↑/k ↓/j move · <space> stag · a stag/unstag all · q quit")
	case focusBranch:
		return helpStyle.Render("↑/k ↓/j move · space stage · enter checkout · n new branch · r rename · d delete · q quit")
	default:
		return ""
	}
}
