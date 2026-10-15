package ui

import (
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

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

	return m.titledPanel(focusStag, "Files", 0, p.height, strings.Join(s, "\n"))
}
