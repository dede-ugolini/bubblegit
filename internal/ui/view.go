package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// visibleWindow returns the [start, end) bounds of a scrolling window over
// n items in a panel of the given rendered height. Border top+bottom eat 2
// rows; without this, a list bigger than the panel's height renders every
// entry and grows the box past height, throwing off the rest of the
// layout. The window is centered on idx (clamped to the list's bounds) so
// the cursor stays in view as it moves.
func visibleWindow(n, height, idx int) (start, end int) {
	visible := height - 2
	if visible < 1 {
		visible = 1
	}

	if n > visible {
		start = idx - visible/2
		if start < 0 {
			start = 0
		}
		if start > n-visible {
			start = n - visible
		}
	}
	end = start + visible
	if end > n {
		end = n
	}
	return start, end
}

// panelAt maps a screen cell (x, y), as reported by a mouse event, to which
// panel it falls in and - for the four list panels - which item row,
// following the same visibleWindow scroll math the renderers use to lay
// out those rows. row is -1 when the cell is on the panel's border or
// below its last item (still a hit on the panel, just not on a row). ok is
// false when the cell isn't inside any panel at all (footer, gaps).
//
// This mirrors renderNormalView's layout by construction: the four list
// panels stack top-to-bottom in a left column of shared width, with the
// diff panel to their right starting where that column ends - so a panel
// resize (including the panelFullScreen zeroing-out of the rest) is picked
// up for free from the same Height/Width fields the renderers already use.
func (m Model) panelAt(x, y int) (target, row int, ok bool) {
	type listPanel struct {
		focus         int
		height, width int
		idx, count    int
	}
	panels := []listPanel{
		{focusStag, m.filesHeight, m.filesWidth, m.idxFiles, len(m.files)},
		{focusBranch, m.branchHeight, m.branchWidth, m.idxBranch, len(m.branches)},
		{focusLog, m.logHeight, m.logWidth, m.idxLog, len(m.log)},
		{focusStash, m.stashHeight, m.stashWidth, m.idxStash, len(m.stashes)},
	}

	top := 0
	maxWidth := 0
	for _, p := range panels {
		if p.width > maxWidth {
			maxWidth = p.width
		}
		if p.height <= 0 || p.width <= 0 {
			continue
		}
		if x < 0 || x >= p.width+2 || y < top || y >= top+p.height {
			top += p.height
			continue
		}
		if y == top || y == top+p.height-1 {
			// Border row: still a hit on the panel, no row under it.
			return p.focus, -1, true
		}
		start, _ := visibleWindow(p.count, p.height, p.idx)
		row = start + (y - top - 1)
		if row >= p.count {
			row = -1
		}
		return p.focus, row, true
	}

	diffX := 0
	if maxWidth > 0 {
		diffX = maxWidth + 2
	}
	if m.diff.Height() > 0 && m.diff.Width() > 0 &&
		x >= diffX && x < diffX+m.diff.Width()+2 &&
		y >= 0 && y < m.diff.Height() {
		return focusDiff, -1, true
	}

	return 0, -1, false
}

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

	if m.tagPopup.active {
		popupWidth := m.width / 4
		popupHeight := m.height / 4

		popup := lipgloss.NewLayer(m.renderTagPopup()).
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

	if m.mergePopup.active {
		popupWidth := m.width / 3
		popupHeight := m.height / 5

		popup := lipgloss.NewLayer(m.renderMergePopup()).
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

	if m.stashBranchPopup.active {
		popupWidth := m.width / 3
		popupHeight := m.height / 8

		popup := lipgloss.NewLayer(m.renderStashBranchPopup()).
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

	if m.dropConfirm {
		popupWidth := m.width / 3
		popupHeight := m.height / 8

		popup := lipgloss.NewLayer(m.renderDropConfirmPopup()).
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

	if m.deleteBranchConfirm {
		popupWidth := m.width / 3
		popupHeight := m.height / 8

		popup := lipgloss.NewLayer(m.renderDeleteBranchConfirmPopup()).
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

	if m.restoreFileConfirm {
		popupWidth := m.width / 3
		popupHeight := m.height / 8

		popup := lipgloss.NewLayer(m.renderRestoreFileConfirmPopup()).
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

	if m.dropStashConfirm {
		popupWidth := m.width / 3
		popupHeight := m.height / 8

		popup := lipgloss.NewLayer(m.renderDropStashConfirmPopup()).
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

	if m.inputPopup.active {
		popupWidth := m.width / 3
		popupHeight := m.height / 8

		popup := lipgloss.NewLayer(m.renderInputPopup()).
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
	help := lipgloss.NewStyle().Foreground(m.theme.Muted).
		Render("y/enter confirm · n/esc cancel")

	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.FocusBorder).
		Render("Remove all stash entries?\n\n" + help)
}

func (m *Model) renderDropConfirmPopup() string {
	help := lipgloss.NewStyle().Foreground(m.theme.Muted).
		Render("y/enter confirm · n/esc cancel")

	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.FocusBorder).
		Render("Drop last commit?\n\n" + m.dropSubject + "\n\n" + help)
}

func (m *Model) renderDeleteBranchConfirmPopup() string {
	help := lipgloss.NewStyle().Foreground(m.theme.Muted).
		Render("y/enter confirm · n/esc cancel")

	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.FocusBorder).
		Render("Delete branch '" + m.deleteBranchName + "'?\n\n" + help)
}

func (m *Model) renderRestoreFileConfirmPopup() string {
	help := lipgloss.NewStyle().Foreground(m.theme.Muted).
		Render("y/enter confirm · n/esc cancel")

	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.FocusBorder).
		Render("Restore '" + m.restoreFilePath + "'?\n\n" + help)
}

func (m *Model) renderDropStashConfirmPopup() string {
	help := lipgloss.NewStyle().Foreground(m.theme.Muted).
		Render("y/enter confirm · n/esc cancel")

	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.FocusBorder).
		Render("Drop stash '" + m.dropStashMessage + "'?\n\n" + help)
}

func (m *Model) renderStashBranchPopup() string {
	help := lipgloss.NewStyle().Foreground(m.theme.Muted).
		Render("enter confirm · esc cancel")

	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.FocusBorder).
		Render("Branch from stash\n\n" + m.stashBranchPopup.input.View() + "\n" + help)
}

func (m *Model) renderMergePopup() string {
	help := lipgloss.NewStyle().Foreground(m.theme.Muted).
		Render("↑/k ↓/j select · enter confirm · esc cancel")

	var lines []string
	lines = append(lines, "Merge '"+m.mergePopup.branch+"' into current branch:")
	for i, mode := range mergeModes {
		prefix := "  "
		style := lipgloss.NewStyle().Width(m.width / 3)
		if m.mergePopup.idx == i {
			prefix = "› "
			style = style.Background(m.theme.Accent)
		}
		lines = append(lines, style.Render(prefix+mode.String()))
	}

	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.FocusBorder).
		Render(strings.Join(lines, "\n") + "\n\n" + help)
}

func (m *Model) renderInputPopup() string {
	help := lipgloss.NewStyle().Foreground(m.theme.Muted).
		Render("enter confirm · esc cancel")

	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.FocusBorder).
		Render(m.inputPopup.title + "\n\n" + m.inputPopup.input.View() + "\n" + help)
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

func (m *Model) renderTagPopup() string {
	var name, message string
	width := m.width / 3

	name = lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		Render(m.tagPopup.tagName.View())

	message = lipgloss.NewStyle().
		Width(width).
		Height(m.height / 4).
		Border(lipgloss.RoundedBorder()).
		Render(m.tagPopup.tagMessage.View())

	return name + "\n" + message
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

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(m.theme.Error).Render(m.err.Error()))
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left, b.String(), m.renderDiff(),
	) + lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render("\n\n"+m.renderFooter())
}

func (m *Model) renderDiff() string {
	if m.focus == focusDiff {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(m.theme.FocusBorder).
			Bold(true).
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
	red := lipgloss.NewStyle().Foreground(m.theme.Removed)
	green := lipgloss.NewStyle().Foreground(m.theme.Added)
	yellow := lipgloss.NewStyle().Foreground(m.theme.Conflict)

	start, end := visibleWindow(len(m.files), m.filesHeight, m.idxFiles)
	for i := start; i < end; i++ {
		f := m.files[i]
		stag := string(f.Index)
		worktree := string(f.Worktree)
		path := f.Path

		if i == m.idxFiles && m.focus == focusStag {
			// Deliberately plain text here: coloring stag/worktree/path
			// individually first and only then wrapping the joined line in
			// Width(...).Background(...) would hit the same nested-reset
			// issue as log/stash - see there.
			s = append(s, lipgloss.NewStyle().Width(m.filesWidth).Background(m.theme.Accent).Render(stag+worktree+" "+path))
			continue
		}

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

		s = append(s, lipgloss.NewStyle().Width(m.filesWidth).Render(stag+worktree+" "+path))
	}

	// The outer style deliberately has no Width of its own - see the note
	// on renderStash.
	if len(s) == 0 {
		s = append(s, lipgloss.NewStyle().Width(m.filesWidth).Render(""))
	}

	if m.focus == focusStag {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(m.theme.FocusBorder).
			Bold(true).
			Height(m.filesHeight).
			Render(strings.Join(s, "\n"))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Height(m.filesHeight).
		Render(strings.Join(s, "\n"))
}

func (m *Model) renderBranches() string {
	if m.branchHeight <= 0 || m.branchWidth <= 0 {
		return ""
	}
	var names []string

	start, end := visibleWindow(len(m.branches), m.branchHeight, m.idxBranch)
	for i := start; i < end; i++ {
		b := m.branches[i]
		if m.idxBranch == i && m.focus == focusBranch {
			// Deliberately plain text here: coloring the current branch's
			// name first and only then wrapping in Width(...).Background(...)
			// would hit the same nested-reset issue as log/stash - see
			// there.
			names = append(names, lipgloss.NewStyle().Width(m.branchWidth).Background(m.theme.Cursor).Render(b.Name))
			continue
		}
		name := b.Name
		if b.Current {
			name = lipgloss.NewStyle().Foreground(m.theme.Accent).Render(name)
		}
		names = append(names, lipgloss.NewStyle().Width(m.branchWidth).Render(name))
	}

	// The outer style deliberately has no Width of its own - see the note
	// on renderStash.
	if len(names) == 0 {
		names = append(names, lipgloss.NewStyle().Width(m.branchWidth).Render(""))
	}

	if m.focus == focusBranch {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(m.theme.FocusBorder).
			Bold(true).
			Height(m.branchHeight).
			Render(strings.Join(names, "\n"))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Height(m.branchHeight).
		Render(strings.Join(names, "\n"))
}

func (m *Model) renderLog() string {
	if m.logHeight <= 0 || m.logWidth <= 0 {
		return ""
	}

	start, end := visibleWindow(len(m.log), m.logHeight, m.idxLog)
	dateColor := lipgloss.NewStyle().Foreground(m.theme.Date)
	shortHashColor := lipgloss.NewStyle().Foreground(m.theme.Hash)

	lo, hi := m.squashAnchor, m.idxLog
	if lo > hi {
		lo, hi = hi, lo
	}

	var entrys []string
	for i := start; i < end; i++ {
		l := m.log[i]
		if (i == m.idxLog && m.focus == focusLog) || (m.squashMarking && i >= lo && i <= hi) {
			// Deliberately plain text here: dateColor/shortHashColor each
			// end in their own ANSI reset, which - nested inside this
			// Background() - would wipe the highlight out from under the
			// date and subject the moment it's hit (\x1b[m clears every
			// SGR attribute, not just foreground).
			entrys = append(entrys, lipgloss.NewStyle().Width(m.logWidth).Background(m.theme.Cursor).Render(l.ShortHash+" "+l.Date+" "+l.Subject))
			continue
		}
		date := dateColor.Render(l.Date)
		shortHash := shortHashColor.Render(l.ShortHash)
		entrys = append(entrys, lipgloss.NewStyle().Width(m.logWidth).Render(shortHash+" "+date+" "+l.Subject))
	}
	// The outer style deliberately has no Width of its own - see the
	// matching note in renderStash below.
	if len(entrys) == 0 {
		entrys = append(entrys, lipgloss.NewStyle().Width(m.logWidth).Render(""))
	}
	if m.focus == focusLog {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(m.theme.FocusBorder).
			Bold(true).
			Height(m.logHeight).
			Render(strings.Join(entrys, "\n"))
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Height(m.logHeight).
		Render(strings.Join(entrys, "\n"))
}

func (m *Model) renderStash() string {
	if m.stashHeight <= 0 || m.stashWidth <= 0 {
		return ""
	}

	start, end := visibleWindow(len(m.stashes), m.stashHeight, m.idxStash)
	dateColor := lipgloss.NewStyle().Foreground(m.theme.Date)
	refColor := lipgloss.NewStyle().Foreground(m.theme.Hash)

	var entrys []string
	for i := start; i < end; i++ {
		s := m.stashes[i]
		if i == m.idxStash && m.focus == focusStash {
			// Deliberately plain text here: dateColor/refColor each end in
			// their own ANSI reset, which - nested inside this Background()
			// - would wipe the highlight out from under the date and
			// message the moment it's hit (\x1b[m clears every SGR
			// attribute, not just foreground).
			line := s.Ref + " " + s.Date + " " + s.Message
			entrys = append(entrys, lipgloss.NewStyle().Width(m.stashWidth).Background(m.theme.Cursor).Render(line))
			continue
		}
		date := dateColor.Render(s.Date)
		ref := refColor.Render(s.Ref)
		line := ref + " " + date + " " + s.Message
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
			BorderForeground(m.theme.FocusBorder).
			Bold(true).
			Height(m.stashHeight).
			Render(strings.Join(entrys, "\n"))
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Height(m.stashHeight).
		Render(strings.Join(entrys, "\n"))
}

func (m *Model) renderFooter() string {
	helpStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
	mode := "delta"
	if !m.useDelta {
		mode = "git"
	}
	modeStr := helpStyle.Render("diff: " + mode + " · V toggle")
	switch m.focus {
	case focusStag:
		return helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · <space> stag · a stag/unstag all · d restore · t theme: %s · q quit", m.themeName)) + " · " + modeStr
	case focusBranch:
		return helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · enter checkout · n new branch · r rename · d delete · M merge · P push · R set remote · t theme: %s · q quit", m.themeName)) + " · " + modeStr
	case focusLog:
		return helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · pgup/pgdown scroll · r reword · S squash · esc cancel · d drop last · T tag · t theme: %s · q quit", m.themeName)) + " · " + modeStr
	case focusStash:
		return helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · enter apply · n new stash · p pop · b branch · d drop · D clear all · t theme: %s · q quit", m.themeName)) + " · " + modeStr
	case focusDiff:
		return helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · pgup/pgdown scroll · t theme: %s · q quit", m.themeName)) + " · " + modeStr
	default:
		return helpStyle.Render("t theme · q quit") + " · " + modeStr
	}
}
