package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

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
	return m.titledPanel(focusLog, "Log", 0, p.height, strings.Join(entrys, "\n"))
}
