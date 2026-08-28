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

	if m.commitPopup.active {
		popupWidth := m.width / 4
		popupHeight := m.height / 4

		popup := lipgloss.NewLayer(m.renderPopup()).
			X((m.width - popupWidth) / 2).
			Y((m.height - popupHeight) / 2).
			Z(1)

		base := lipgloss.NewLayer(m.renderNormalView()).
			Z(0)

		c := lipgloss.NewCompositor(base, popup)
		v := tea.NewView(c.Render())
		v.AltScreen = true
		v.MouseMode = tea.MouseModeCellMotion
		return v
	}

	if m.stashClearConfirm {
		popupWidth := m.width / 3
		popupHeight := m.height / 8

		popup := lipgloss.NewLayer(m.renderStashClearConfirmPopup()).
			X((m.width - popupWidth) / 2).
			Y((m.height - popupHeight) / 2).
			Z(1)

		base := lipgloss.NewLayer(m.renderNormalView()).
			Z(0)

		c := lipgloss.NewCompositor(base, popup)
		v := tea.NewView(c.Render())
		v.AltScreen = true
		v.MouseMode = tea.MouseModeCellMotion
		return v
	}

	v := tea.NewView(m.renderNormalView())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m *Model) renderStashClearConfirmPopup() string {
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
		Render("y/enter confirm · n/esc cancel")

	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.ANSIColor(222)).
		Render("Remove all stash entries?\n\n" + help)
}

func (m *Model) renderPopup() string {
	var summary, message string
	width := m.width / 3

	summary = lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		Render(m.commitPopup.commitSummary.View())

	message = lipgloss.NewStyle().
		Width(width).
		Height(m.height / 4).
		Border(lipgloss.RoundedBorder()).
		Render(m.commitPopup.commitMessage.View())

	return summary + "\n" + message
}

func (m *Model) renderNormalView() string {
	var b strings.Builder

	b.WriteString(m.renderFiles())
	b.WriteString("\n")

	b.WriteString(m.renderBranches())
	b.WriteString("\n")

	b.WriteString(m.renderLog())
	b.WriteString("\n")

	b.WriteString(m.renderStash())

	if m.focusInput {
		b.WriteString(m.input.View())
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(m.err.Error()))
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left, b.String(), m.renderDiff(),
	) + "\n" + renderFooter(m.focus)
}

func (m *Model) renderDiff() string {
	if m.focus == focusDiff {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(lipgloss.ANSIColor(222)).
			Height(m.diff.Height()).
			Width(m.diff.Width()).
			Render(m.diff.View())
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Height(m.diff.Height()).
		Width(m.diff.Width()).
		Render(m.diff.View())
}

func (m *Model) renderFiles() string {
	if m.filesHeight <= 0 || m.filesWidth <= 0 {
		return ""
	}
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
	if m.branchHeight <= 0 || m.branchWidth <= 0 {
		return ""
	}
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
	// Border top+bottom eat 2 rows; without windowing, a log limit bigger
	// than the panel's height renders every entry and grows the box past
	// m.logHeight, throwing off the rest of the layout.
	if m.logHeight <= 0 || m.logWidth <= 0 {
		return ""
	}

	visible := m.logHeight - 2
	if visible < 1 {
		visible = 1
	}

	start := 0
	if len(m.log) > visible {
		start = m.idxLog - visible/2
		if start < 0 {
			start = 0
		}
		if start > len(m.log)-visible {
			start = len(m.log) - visible
		}
	}
	end := start + visible
	if end > len(m.log) {
		end = len(m.log)
	}

	var entrys []string
	for i := start; i < end; i++ {
		l := m.log[i]
		if i == m.idxLog && m.focus == focusLog {
			entrys = append(entrys, lipgloss.NewStyle().Width(m.logWidth).Background(lipgloss.ANSIColor(9)).Render(l.ShortHash+" "+l.Date+" "+l.Subject))
			continue
		}
		entrys = append(entrys, lipgloss.NewStyle().Width(m.logWidth).Render(l.ShortHash+" "+l.Date+" "+l.Subject))
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

func (m *Model) renderStash() string {
	if m.stashHeight <= 0 || m.stashWidth <= 0 {
		return ""
	}

	visible := m.stashHeight - 2
	if visible < 1 {
		visible = 1
	}

	start := 0
	if len(m.stashes) > visible {
		start = m.idxStash - visible/2
		if start < 0 {
			start = 0
		}
		if start > len(m.stashes)-visible {
			start = len(m.stashes) - visible
		}
	}
	end := start + visible
	if end > len(m.stashes) {
		end = len(m.stashes)
	}

	var entrys []string
	for i := start; i < end; i++ {
		s := m.stashes[i]
		line := s.Ref + " " + s.Date + " " + s.Message
		if i == m.idxStash && m.focus == focusStash {
			entrys = append(entrys, lipgloss.NewStyle().Width(m.stashWidth).Background(lipgloss.ANSIColor(9)).Render(line))
			continue
		}
		entrys = append(entrys, lipgloss.NewStyle().Width(m.stashWidth).Render(line))
	}
	// The outer style deliberately has no Width of its own (re-applying
	// Width on top of an already width-padded, already-styled line
	// corrupts its background - see the highlighted branch below). That
	// means an empty list has nothing to establish the panel's width, so
	// the border collapses to zero. Pad a single blank line to hold it.
	if len(entrys) == 0 {
		entrys = append(entrys, lipgloss.NewStyle().Width(m.stashWidth).Render(""))
	}
	if m.focus == focusStash {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(lipgloss.ANSIColor(222)).
			Height(m.stashHeight).
			Render(strings.Join(entrys, "\n"))
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Height(m.stashHeight).
		Render(strings.Join(entrys, "\n"))
}

func renderFooter(focus int) string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	switch focus {
	case focusStag:
		return helpStyle.Render("↑/k ↓/j move · <space> stag · a stag/unstag all · d restore · 0 diff · 1 files · 2 branches · 3 log · 4 stash · q quit")
	case focusBranch:
		return helpStyle.Render("↑/k ↓/j move · enter checkout · n new branch · r rename · d delete · 0 diff · 1 files · 2 branches · 3 log · 4 stash · q quit")
	case focusLog:
		return helpStyle.Render("↑/k ↓/j move · pgup/pgdown scroll · 0 diff · 1 files · 2 branches · 3 log · 4 stash · q quit")
	case focusStash:
		return helpStyle.Render("↑/k ↓/j move · enter apply · n new stash · d drop · D clear all · 0 diff · 1 files · 2 branches · 3 log · 4 stash · q quit")
	case focusDiff:
		return helpStyle.Render("↑/k ↓/j move · pgup/pgdown scroll · 0 diff · 1 files · 2 branches · 3 log · 4 stash · q quit")
	default:
		return helpStyle.Render("0 diff · 1 files · 2 branches · 3 log · 4 stash · q quit")
	}
}
