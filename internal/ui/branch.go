package ui

import (
	"fmt"
	"strings"
	"time"

	"bubblegit/internal/git"

	"charm.land/lipgloss/v2"
)

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

	return m.titledPanel(focusBranch, "Branches", 0, p.height, strings.Join(names, "\n"))
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
