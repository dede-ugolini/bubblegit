package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme is every color decision the UI makes, gathered in one place so a
// color's meaning has a name instead of leaving readers to remember which
// magic index/hex means what, and so a whole theme is just one more value
// of this type.
type Theme struct {
	// FocusBorder marks the border of whichever panel currently holds
	// keyboard focus.
	FocusBorder color.Color

	// Cursor highlights the selected row within a focused list panel
	// (branches, log, stash).
	Cursor color.Color

	// Accent marks something as "this one": the checked-out branch's name,
	// and the single-selection cursor in the files panel.
	Accent color.Color

	// Added, Removed, StagedModified and Conflict color file status
	// markers: staged, unstaged/untracked, staged-but-since-modified, and
	// unmerged - the last being a path `git commit` will refuse until it is
	// resolved.
	Added          color.Color
	Removed        color.Color
	StagedModified color.Color
	Conflict       color.Color

	// Error renders the error banner text.
	Error color.Color

	// Muted is for de-emphasized text: footer/help hints.
	Muted color.Color

	// Date colors the date field in the log/stash panels.
	Date color.Color

	// Hash colors the commit short-hash / stash ref field in the log/stash
	// panels.
	Hash color.Color

	// Ahead colors the short hash of a log entry whose commit the current
	// branch carries but its upstream doesn't (push would publish it).
	Ahead color.Color

	// Author colors the author initials shown in the log panel.
	Author color.Color
}

// themeSystem relies entirely on the terminal's own ANSI palette (basic
// 4-bit colors plus a couple of 256-color grays), so it inherits whatever
// light/dark scheme the user already has configured instead of imposing
// one.
var themeSystem = Theme{
	FocusBorder:    lipgloss.ANSIColor(222),
	Cursor:         lipgloss.ANSIColor(9),
	Accent:         lipgloss.ANSIColor(11),
	Added:          lipgloss.ANSIColor(lipgloss.Green),
	Removed:        lipgloss.ANSIColor(lipgloss.Red),
	StagedModified: lipgloss.ANSIColor(lipgloss.Yellow),
	Conflict:       lipgloss.ANSIColor(lipgloss.Red),
	Error:          lipgloss.ANSIColor(lipgloss.Red),
	Muted:          lipgloss.Color("240"),
	Date:           lipgloss.ANSIColor(12),
	Hash:           lipgloss.ANSIColor(14),
	Ahead:          lipgloss.ANSIColor(lipgloss.Red),
	Author:         lipgloss.ANSIColor(13),
}

// themeNord maps the same slots onto the Nord palette
// (https://www.nordtheme.com/), fixed hex colors instead of terminal ANSI
// slots.
var themeNord = Theme{
	FocusBorder:    lipgloss.Color("#88C0D0"), // nord8  - frost, light cyan
	Cursor:         lipgloss.Color("#5E81AC"), // nord10 - frost, dark blue
	Accent:         lipgloss.Color("#EBCB8B"), // nord13 - aurora, yellow
	Added:          lipgloss.Color("#A3BE8C"), // nord14 - aurora, green
	Removed:        lipgloss.Color("#BF616A"), // nord11 - aurora, red
	StagedModified: lipgloss.Color("#D08770"), // nord12 - aurora, orange
	Conflict:       lipgloss.Color("#BF616A"), // nord11 - aurora, red
	Error:          lipgloss.Color("#BF616A"), // nord11 - aurora, red
	Muted:          lipgloss.Color("#4C566A"), // nord3  - polar night, gray
	Date:           lipgloss.Color("#81A1C1"), // nord9  - frost, blue
	Hash:           lipgloss.Color("#8FBCBB"), // nord7  - frost, teal
	Ahead:          lipgloss.Color("#A3BE8C"), // nord14 - aurora, green
	Author:         lipgloss.Color("#B48EAD"), // nord15 - frost, purple
}

// themeOrder is the cycling order for the "t" keybind; themesByName must have
// exactly these keys.
var themeOrder = []string{"system", "nord"}

var themesByName = map[string]Theme{
	"system": themeSystem,
	"nord":   themeNord,
}

// nextTheme returns the name and value of the theme that follows current
// in themeOrder, wrapping around. An unrecognized current name (there
// isn't one today, but NewModel shouldn't have to know that) falls back to
// the first theme.
func nextTheme(current string) (string, Theme) {
	for i, name := range themeOrder {
		if name == current {
			next := themeOrder[(i+1)%len(themeOrder)]
			return next, themesByName[next]
		}
	}
	return themeOrder[0], themesByName[themeOrder[0]]
}
