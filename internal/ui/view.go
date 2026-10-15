package ui

import (
	"fmt"
	"image/color"
	"strings"

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
// resize (including the layouts that zero out every other panel) is picked
// up for free from the same Height/Width fields the renderers already use.
// The diff panel is the exception, since its viewport's width and height
// mean different things; see below.
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
	// The diff panel's two numbers mean different things in each direction,
	// which is the one place panelAt can't mirror the list panels' arithmetic
	// directly. m.diff.Width() is the whole box, since renderDiff pads the
	// viewport out to it, so no border has to be added; m.diff.Height() is
	// the viewport inside the box, so the box is two rows taller. Adding the
	// borders to the width made the hit box reach past the right edge, and
	// omitting them from the height left the box's bottom rows unclickable.
	if m.diff.Height() > 0 && m.diff.Width() > 0 &&
		x >= diffX && x < diffX+m.diff.Width() &&
		y >= 0 && y < m.diff.Height()+2 {
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
	case m.worktreePopup.active:
		return m.overlay(m.renderWorktreePopup(), 3, 8)
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

// renderWorktreePopup draws the new-worktree prompt: a path field and an
// optional branch field. The focused field is just the textinput's own
// cursor, so no extra highlight is needed.
func (m *Model) renderWorktreePopup() string {
	help := lipgloss.NewStyle().Foreground(m.theme.Muted).
		Render("tab switch · ctrl+s add · esc cancel")

	return lipgloss.NewStyle().
		Width(m.width / 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.FocusBorder).
		Render("New worktree\n\n" +
			m.worktreePopup.path.View() + "\n" +
			m.worktreePopup.branch.View() + "\n" + help)
}

func (m *Model) renderNormalView() string {
	// The four list panels are joined with newlines, but a layout that
	// collapses one leaves it rendering as the empty string - and a blank
	// line would then take its place, pushing every panel below it down the
	// screen. So the separators follow the panels that are actually there.
	rows := []string{
		m.renderView(focusStag),
		m.renderView(focusBranch),
		m.renderView(focusLog),
		m.renderView(focusStash),
	}

	visible := rows[:0]
	for _, r := range rows {
		if r != "" {
			visible = append(visible, r)
		}
	}
	b := strings.Join(visible, "\n")

	if m.err != nil {
		b += "\n" + lipgloss.NewStyle().Foreground(m.theme.Error).Render(m.err.Error())
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left, b, m.renderDiff(),
	) + lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render("\n\n"+m.renderFooter())
}

func (m *Model) renderDiff() string {
	// Same guard the list panels have: a layout that collapses the diff still
	// needs to render nothing, and without this titledPanel would wrap the
	// empty viewport in a degenerate three-line box.
	if m.diff.Height() <= 0 || m.diff.Width() <= 0 {
		return ""
	}
	// Width includes the borders, so m.diff.Width() is the whole panel and
	// the viewport is padded to it.
	return m.titledPanel(focusDiff, "Diff", m.diff.Width(), m.diff.Height(), m.diff.View())
}

// panelBorder is the border every list panel and the diff panel is drawn in,
// matching the popups. titledPanel takes its corners and fill rune from here
// so the title edge can't drift away from the box it belongs to.
var panelBorder = lipgloss.RoundedBorder()

// titledPanel renders content in the panel's border with title drawn into the
// top edge:
//
//	╭─ Files ─────────╮
//	│ ...             │
//	╰─────────────────╯
//
// lipgloss v2 has no border-title support, and it offers no way to draw only
// three of the four edges either - BorderTop(false) still reserves the row,
// so a panel would come out a line taller. So the box is rendered exactly as
// it was and its top line is then replaced, which leaves a panel's height and
// width untouched and visibleWindow and panelAt needing no adjustment.
// Measuring the line being replaced is what keeps the title edge exactly as
// wide as the box it belongs to, whatever the content happened to be.
//
// bodyWidth is the Width to set on the body style, or 0 to leave it unset:
// the diff panel needs one so its viewport is padded, but the list panels must
// go without it, since re-applying Width to already-padded, already-styled
// lines corrupts their backgrounds (see the note on renderStash).
//
// A title too wide for the panel is dropped rather than allowed to widen or
// misalign the box.
func (m *Model) titledPanel(focus int, title string, bodyWidth, height int, content string) string {
	body := lipgloss.NewStyle().
		Border(panelBorder, true).
		Height(height)
	if bodyWidth > 0 {
		body = body.Width(bodyWidth)
	}
	if m.focus == focus {
		body = body.BorderForeground(m.theme.FocusBorder).Bold(true)
	}

	lines := strings.Split(body.Render(content), "\n")
	if len(lines) == 0 || lines[0] == "" {
		// A panel collapsed by fullscreen renders nothing at all; don't give
		// it a title edge it has no box to sit on.
		return strings.Join(lines, "\n")
	}
	width := lipgloss.Width(lines[0])

	// panelBorder's fill rune and corners around "Title ", then enough fill
	// to reach the width of the line being replaced. The edge is colored with
	// the same color lipgloss gave the border, and left unbolded, because Bold
	// reaches a panel's content but not its border glyphs.
	label := panelBorder.Top + " " + title + " "
	fill := width - 2 - lipgloss.Width(label)
	if lipgloss.Width(label)+2 > width {
		// Too narrow for the title: keep the box's width and drop the
		// title rather than overflowing.
		label, fill = "", width-2
	}
	if fill < 0 {
		fill = 0
	}
	edge := lipgloss.NewStyle()
	if m.focus == focus {
		edge = edge.Foreground(m.theme.FocusBorder)
	}
	lines[0] = edge.Render(panelBorder.TopLeft + label +
		strings.Repeat(panelBorder.Top, fill) + panelBorder.TopRight)

	return strings.Join(lines, "\n")
}

type fileIcon struct {
	r rune
	c color.Color
}

// renderView renders whichever view a list panel is currently showing. It
// dispatches on the panel's active view first, then falls back to the
// panel's own renderer, so adding a subpanel only means adding a case here
// and in viewRows.
func (m *Model) renderView(focus int) string {
	switch m.panels[focus].view {
	case viewWorktrees:
		return m.renderWorktrees()
	}
	switch focus {
	case focusStag:
		return m.renderFiles()
	case focusBranch:
		return m.renderBranches()
	case focusLog:
		return m.renderLog()
	case focusStash:
		return m.renderStash()
	}
	return ""
}

func (m *Model) renderFooter() string {
	helpStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
	mode := "delta"
	if !m.useDelta {
		mode = "git"
	}
	// Detached HEAD is global state, not per-panel, so it prefixes every
	// footer's hints rather than being folded into one focus branch. It goes
	// first because it is the one thing that changes what the other hints
	// mean: while detached, "enter checkout" in the log panel is how the
	// working tree got here, and the branch panel is the way back.
	detached := ""
	if m.detached != "" {
		detached = lipgloss.NewStyle().Foreground(m.theme.Error).Render(
			"detached HEAD at " + m.detached + " · ")
	}
	modeStr := helpStyle.Render("diff: " + mode + " · V toggle")
	switch m.focus {
	case focusStag:
		if m.panels[focusStag].view == viewWorktrees {
			return detached + helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · n new worktree · d remove · w files · t theme: %s · q quit", m.themeName)) + " · " + modeStr
		}
		return detached + helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · <space> stag · a stag/unstag all · d restore · w worktrees · t theme: %s · q quit", m.themeName)) + " · " + modeStr
	case focusBranch:
		return detached + helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · enter checkout · n new branch · r rename · d delete · M merge · P push · R set remote · t theme: %s · q quit", m.themeName)) + " · " + modeStr
	case focusLog:
		return detached + helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · enter checkout · pgup/pgdown scroll · r reword · S squash · esc cancel · d drop last · T tag · t theme: %s · q quit", m.themeName)) + " · " + modeStr
	case focusStash:
		return detached + helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · enter apply · n new stash · p pop · b branch · d drop · D clear all · t theme: %s · q quit", m.themeName)) + " · " + modeStr
	case focusDiff:
		return detached + helpStyle.Render(fmt.Sprintf("↑/k ↓/j move · pgup/pgdown scroll · t theme: %s · q quit", m.themeName)) + " · " + modeStr
	default:
		return detached + helpStyle.Render("t theme · q quit") + " · " + modeStr
	}
}
