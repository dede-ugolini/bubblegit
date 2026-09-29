package ui

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"
	"time"

	"bubblegit/internal/git"

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
func (m *Model) panelAt(x, y int) (target, row int, ok bool) {
	// The list panels are the contiguous range focusStag..focusStash, so the
	// slice index doubles as the focus constant.
	top := 0
	maxWidth := 0
	for focus := focusStag; focus < focusDiff; focus++ {
		p := m.panels[focus]
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
			return focus, -1, true
		}
		count := m.rowCount(focus)
		start, _ := visibleWindow(count, p.height, p.idx)
		row = start + (y - top - 1)
		if row >= count {
			row = -1
		}
		return focus, row, true
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

func (m *Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	// TODO: add spinner
	if !m.ready {
		return tea.NewView("Loading...")
	}

	// Popups are checked in order and at most one is ever open, since each
	// one swallows input while it's up. The order below is therefore
	// presentation-only, not a precedence rule.
	switch {
	case m.commitPopup.active:
		return m.overlay(m.renderPopup(), 4, 4)
	case m.tagPopup.active:
		return m.overlay(m.renderTagPopup(), 4, 4)
	case m.mergePopup.active:
		return m.overlay(m.renderMergePopup(), 3, 5)
	case m.stashBranchPopup.active:
		return m.overlay(m.renderStashBranchPopup(), 3, 8)
	case m.confirm.active:
		return m.overlay(m.renderConfirm(), 3, 8)
	case m.inputPopup.active:
		return m.overlay(m.renderInputPopup(), 3, 8)
	}

	v := tea.NewView(m.renderNormalView())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// overlay centers a popup on top of the normal UI. wDiv and hDiv are the
// fractions of the window the popup is assumed to occupy: they decide
// where it is placed, not how big it is, since each renderXxx sizes
// itself.
func (m *Model) overlay(content string, wDiv, hDiv int) tea.View {
	popup := lipgloss.NewLayer(content).
		X((m.width - m.width/wDiv) / 2).
		Y((m.height - m.height/hDiv) / 2).
		Z(1)
	base := lipgloss.NewLayer(m.renderNormalView()).Z(0)
	v := tea.NewView(lipgloss.NewCompositor(base, popup).Render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// renderConfirm draws the open confirmation prompt: the question, its
// optional detail, and the key hint.
func (m *Model) renderConfirm() string {
	c := &m.confirm
	help := c.help
	if help == "" {
		help = defaultConfirmHelp
	}
	body := c.title
	if c.detail != "" {
		body += "\n\n" + c.detail
	}
	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.FocusBorder).
		Render(body + "\n\n" + lipgloss.NewStyle().Foreground(m.theme.Muted).Render(help))
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
	for i, mode := range git.MergeModes {
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

type fileIcon struct {
	r rune
	c color.Color
}

func (m *Model) renderFiles() string {
	p := m.panels[focusStag]
	if p.height <= 0 || p.width <= 0 {
		return ""
	}
	var s []string
	red := lipgloss.NewStyle().Foreground(m.theme.Removed)
	green := lipgloss.NewStyle().Foreground(m.theme.Added)
	yellow := lipgloss.NewStyle().Foreground(m.theme.Conflict)
	dirStyle := lipgloss.NewStyle().Foreground(m.theme.Accent)

	start, end := visibleWindow(len(m.treeRows), p.height, m.panels[focusStag].idx)
	for i := start; i < end; i++ {
		row := m.treeRows[i]
		selected := i == m.panels[focusStag].idx && m.focus == focusStag

		if row.isDir {
			glyph := "▾"
			if m.collapsed[row.dir] {
				glyph = "▸"
			}
			label := strings.Repeat("  ", row.depth) + glyph + " " + row.name + "/"
			if selected {
				// Deliberately plain text here: coloring the dir name first
				// and only then wrapping in Width(...).Background(...) would
				// hit the same nested-reset issue as the file rows below.
				s = append(s, lipgloss.NewStyle().Width(p.width).Background(m.theme.Accent).Render(label))
				continue
			}
			s = append(s, lipgloss.NewStyle().Width(p.width).Render(dirStyle.Render(label)))
			continue
		}

		f := m.files[row.fileIdx]
		stag := string(f.Index)
		worktree := string(f.Worktree)
		path := row.name

		ext := strings.ToLower(filepath.Ext(f.Path))
		var icon string
		fileIcon, hasIcon := extIcons[ext]
		if hasIcon {
			icon = string(fileIcon.r)
		}

		// The two-space slot keeps files aligned with the directory glyph
		// column of their ancestors.
		indent := strings.Repeat("  ", row.depth) + "  "

		if selected {
			// Deliberately plain text here: coloring stag/worktree/path
			// individually first and only then wrapping the joined line in
			// Width(...).Background(...) would hit the same nested-reset
			// issue as log/stash - see there. The icon is therefore left
			// uncolored as well - its ANSI reset would wipe the highlight.
			s = append(s, lipgloss.NewStyle().Width(p.width).Background(m.theme.Accent).Render(indent+stag+worktree+" "+icon+" "+path))
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

		if hasIcon {
			icon = lipgloss.NewStyle().Foreground(fileIcon.c).Render(icon)
		}
		s = append(s, lipgloss.NewStyle().Width(p.width).Render(indent+stag+worktree+" "+icon+" "+path))
	}

	// The outer style deliberately has no Width of its own - see the note
	// on renderStash.
	if len(s) == 0 {
		s = append(s, lipgloss.NewStyle().Width(p.width).Render(""))
	}

	if m.focus == focusStag {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(m.theme.FocusBorder).
			Bold(true).
			Height(p.height).
			Render(strings.Join(s, "\n"))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Height(p.height).
		Render(strings.Join(s, "\n"))
}

func (m *Model) renderBranches() string {
	p := m.panels[focusBranch]
	if p.height <= 0 || p.width <= 0 {
		return ""
	}
	var names []string

	start, end := visibleWindow(len(m.branches), p.height, m.panels[focusBranch].idx)
	agoColor := lipgloss.NewStyle().Foreground(m.theme.Muted)
	for i := start; i < end; i++ {
		b := m.branches[i]
		track := branchTrack(b)
		ago := ""
		if a := timeAgo(b.CommitTime, time.Now()); a != "" {
			ago = fmt.Sprintf("%4s ", a)
		}
		if m.panels[focusBranch].idx == i && m.focus == focusBranch {
			// Deliberately plain text here (time included): coloring any of
			// it and only then wrapping in Width(...).Background(...) would
			// hit the nested-reset issue seen in log/stash - see there.
			row := ago + b.Name
			if track != "" {
				row += " " + track
			}
			prefix := "   "
			if b.Current {
				prefix = " * "
			}
			names = append(names, lipgloss.NewStyle().Width(p.width).Background(m.theme.Cursor).Render(prefix+row))
			continue
		}
		name := b.Name
		prefix := "   "
		if b.Current {
			prefix = " * "
			name = lipgloss.NewStyle().Foreground(m.theme.Accent).Render(name)
		}
		row := agoColor.Render(ago) + name
		if track != "" {
			row += " " + track
		}
		names = append(names, lipgloss.NewStyle().Width(p.width).Render(prefix+row))
	}

	// The outer style deliberately has no Width of its own - see the note
	// on renderStash.
	if len(names) == 0 {
		names = append(names, lipgloss.NewStyle().Width(p.width).Render(""))
	}

	if m.focus == focusBranch {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(m.theme.FocusBorder).
			Bold(true).
			Height(p.height).
			Render(strings.Join(names, "\n"))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Height(p.height).
		Render(strings.Join(names, "\n"))
}

// branchTrack renders a branch's ahead/behind status relative to its
// upstream as a compact suffix, e.g. "[ahead 1, behind 2]" or "[gone]".
// Returns "" for branches that are in sync or have no upstream.
func branchTrack(b git.BranchInfo) string {
	var parts []string
	if b.Ahead > 0 {
		parts = append(parts, fmt.Sprintf("%d", b.Ahead))
	}
	if b.Behind > 0 {
		parts = append(parts, fmt.Sprintf("%d", b.Behind))
	}
	if b.Gone {
		parts = append(parts, "gone")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}

// timeAgo renders t relative to now as a compact single-unit age like
// lazygit's: "3m", "5h", "2d", "4w", "6mo" or "1y". A zero time (branch
// with no commits) returns ""; sub-minute and future (clock-skewed) ages
// both read as "1m".
func timeAgo(t, now time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := now.Sub(t)
	if d < time.Minute {
		return "1m"
	}
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d/time.Minute))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d/time.Hour))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d/(24*time.Hour)))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw", int(d/(7*24*time.Hour)))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo", int(d/(30*24*time.Hour)))
	default:
		return fmt.Sprintf("%dy", int(d/(365*24*time.Hour)))
	}
}

// authorInitials returns an up-to-two-letter uppercase abbreviation for an
// author name: the first letter of the first and last word ("John Francis
// Doe" -> "JD"). Single-word names ("Linus") are just "L". Returns "" for an
// empty/whitespace-only author so callers can skip rendering it.
func authorInitials(author string) string {
	words := strings.Fields(author)
	if len(words) == 0 {
		return ""
	}
	up := func(b byte) byte {
		if b >= 'a' && b <= 'z' {
			b -= 'a' - 'A'
		}
		return b
	}
	out := []byte{up(words[0][0])}
	if len(words) > 1 {
		out = append(out, up(words[len(words)-1][0]))
	}
	return string(out)
}

func (m *Model) renderLog() string {
	p := m.panels[focusLog]
	if p.height <= 0 || p.width <= 0 {
		return ""
	}

	start, end := visibleWindow(len(m.log), p.height, m.panels[focusLog].idx)
	dateColor := lipgloss.NewStyle().Foreground(m.theme.Date)
	shortHashColor := lipgloss.NewStyle().Foreground(m.theme.Hash)
	aheadColor := lipgloss.NewStyle().Foreground(m.theme.Ahead)
	tagColor := lipgloss.NewStyle().Foreground(m.theme.Accent)
	authorColor := lipgloss.NewStyle().Foreground(m.theme.Author)

	lo, hi := m.squashAnchor, m.panels[focusLog].idx
	if lo > hi {
		lo, hi = hi, lo
	}

	var entrys []string
	for i := start; i < end; i++ {
		l := m.log[i]
		tagStr := ""
		if names := m.tags[l.Hash]; len(names) > 0 {
			tagStr = "(" + strings.Join(names, ", ") + ")"
		}
		if (i == m.panels[focusLog].idx && m.focus == focusLog) || (m.squashMarking && i >= lo && i <= hi) {
			// Deliberately plain text here: dateColor/shortHashColor each
			// end in their own ANSI reset, which - nested inside this
			// Background() - would wipe the highlight out from under the
			// date and subject the moment it's hit (\x1b[m clears every
			// SGR attribute, not just foreground).
			plainLine := l.ShortHash + " " + l.Date
			if tagStr != "" {
				plainLine += " " + tagStr
			}
			if init := authorInitials(l.Author); init != "" {
				plainLine += " " + init
			}
			entrys = append(entrys, lipgloss.NewStyle().Width(p.width).Background(m.theme.Cursor).Render(plainLine+" "+l.Subject))
			continue
		}
		date := dateColor.Render(l.Date)
		hashStyle := shortHashColor
		if m.ahead[l.Hash] {
			hashStyle = aheadColor
		}
		shortHash := hashStyle.Render(l.ShortHash)
		line := shortHash + " " + date
		if tagStr != "" {
			line += " " + tagColor.Render(tagStr)
		}
		if init := authorInitials(l.Author); init != "" {
			line += " " + authorColor.Render(init)
		}
		entrys = append(entrys, lipgloss.NewStyle().Width(p.width).Render(line+" "+l.Subject))
	}
	// The outer style deliberately has no Width of its own - see the
	// matching note in renderStash below.
	if len(entrys) == 0 {
		entrys = append(entrys, lipgloss.NewStyle().Width(p.width).Render(""))
	}
	if m.focus == focusLog {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(m.theme.FocusBorder).
			Bold(true).
			Height(p.height).
			Render(strings.Join(entrys, "\n"))
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Height(p.height).
		Render(strings.Join(entrys, "\n"))
}

func (m *Model) renderStash() string {
	p := m.panels[focusStash]
	if p.height <= 0 || p.width <= 0 {
		return ""
	}

	start, end := visibleWindow(len(m.stashes), p.height, m.panels[focusStash].idx)
	dateColor := lipgloss.NewStyle().Foreground(m.theme.Date)
	refColor := lipgloss.NewStyle().Foreground(m.theme.Hash)

	var entrys []string
	for i := start; i < end; i++ {
		s := m.stashes[i]
		if i == m.panels[focusStash].idx && m.focus == focusStash {
			// Deliberately plain text here: dateColor/refColor each end in
			// their own ANSI reset, which - nested inside this Background()
			// - would wipe the highlight out from under the date and
			// message the moment it's hit (\x1b[m clears every SGR
			// attribute, not just foreground).
			line := s.Ref + " " + s.Date + " " + s.Message
			entrys = append(entrys, lipgloss.NewStyle().Width(p.width).Background(m.theme.Cursor).Render(line))
			continue
		}
		date := dateColor.Render(s.Date)
		ref := refColor.Render(s.Ref)
		line := ref + " " + date + " " + s.Message
		entrys = append(entrys, lipgloss.NewStyle().Width(p.width).Render(line))
	}
	// The outer style deliberately has no Width of its own (re-applying
	// Width on top of an already width-padded, already-styled line
	// corrupts its background - see the highlighted branch below). That
	// means an empty list has nothing to establish the panel's width, so
	// the border collapses to zero. Pad a single blank line to hold it.
	if len(entrys) == 0 {
		entrys = append(entrys, lipgloss.NewStyle().Width(p.width).Render(""))
	}
	if m.focus == focusStash {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(m.theme.FocusBorder).
			Bold(true).
			Height(p.height).
			Render(strings.Join(entrys, "\n"))
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Height(p.height).
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
