package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

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
	return m.titledPanel(focusStash, "Stash", 0, p.height, strings.Join(entrys, "\n"))
}
