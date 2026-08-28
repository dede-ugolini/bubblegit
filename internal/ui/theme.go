package ui

import "charm.land/lipgloss/v2"

// Color palette. Every UI coloring decision goes through one of these so a
// color's meaning has one name instead of leaving readers to remember which
// magic ANSI index means what, and so changing a color only means changing
// it here.
var (
	// colorFocusBorder marks the border of whichever panel currently holds
	// keyboard focus.
	colorFocusBorder = lipgloss.ANSIColor(222)

	// colorCursor highlights the selected row within a focused list panel
	// (branches, log, stash).
	colorCursor = lipgloss.ANSIColor(9)

	// colorAccent marks something as "this one": the checked-out branch's
	// name, and the single-selection cursor in the files panel.
	colorAccent = lipgloss.ANSIColor(11)

	// colorAdded, colorRemoved and colorConflict color file status markers:
	// staged, unstaged/untracked, and staged-but-since-modified again.
	colorAdded    = lipgloss.ANSIColor(lipgloss.Green)
	colorRemoved  = lipgloss.ANSIColor(lipgloss.Red)
	colorConflict = lipgloss.ANSIColor(lipgloss.Yellow)

	// colorError renders the error banner text.
	colorError = lipgloss.ANSIColor(lipgloss.Red)

	// colorMuted is for de-emphasized text: footer/help hints.
	colorMuted = lipgloss.Color("240")
)
