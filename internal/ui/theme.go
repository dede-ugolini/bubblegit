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
	// Background is the terminal background color for the theme. It is set
	// on the tea.View rather than painted into the rendered string, so the
	// whole window is tinted and text with no explicit background picks it
	// up. nil leaves the terminal's own default in place.
	Background color.Color

	// FocusBorder marks the border of whichever panel currently holds
	// keyboard focus.
	FocusBorder color.Color

	// Cursor highlights the selected row within a focused list panel
	// (branches, log, stash).
	Cursor color.Color

	// Accent marks something as "this one": the checked-out branch's name,
	// and the single-selection cursor in the files panel.
	Accent color.Color

	// Added, Removed and Conflict color file status markers: staged,
	// unstaged/untracked, and staged-but-since-modified.
	Added    color.Color
	Removed  color.Color
	Conflict color.Color

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
	// No Background: this theme exists to defer to whatever light/dark
	// scheme the terminal already uses, so the default background is left
	// untouched.
	FocusBorder: lipgloss.ANSIColor(222),
	Cursor:      lipgloss.ANSIColor(9),
	Accent:      lipgloss.ANSIColor(11),
	Added:       lipgloss.ANSIColor(lipgloss.Green),
	Removed:     lipgloss.ANSIColor(lipgloss.Red),
	Conflict:    lipgloss.ANSIColor(lipgloss.Yellow),
	Error:       lipgloss.ANSIColor(lipgloss.Red),
	Muted:       lipgloss.Color("240"),
	Date:        lipgloss.ANSIColor(12),
	Hash:        lipgloss.ANSIColor(14),
	Ahead:       lipgloss.ANSIColor(lipgloss.Red),
	Author:      lipgloss.ANSIColor(13),
}

// themeNord maps the same slots onto the Nord palette
// (https://www.nordtheme.com/), fixed hex colors instead of terminal ANSI
// slots.
var themeNord = Theme{
	Background:  lipgloss.Color("#2E3440"), // nord0  - polar night, bg
	FocusBorder: lipgloss.Color("#88C0D0"), // nord8  - frost, light cyan
	Cursor:      lipgloss.Color("#5E81AC"), // nord10 - frost, dark blue
	Accent:      lipgloss.Color("#EBCB8B"), // nord13 - aurora, yellow
	Added:       lipgloss.Color("#A3BE8C"), // nord14 - aurora, green
	Removed:     lipgloss.Color("#BF616A"), // nord11 - aurora, red
	Conflict:    lipgloss.Color("#D08770"), // nord12 - aurora, orange
	Error:       lipgloss.Color("#BF616A"), // nord11 - aurora, red
	Muted:       lipgloss.Color("#4C566A"), // nord3  - polar night, gray
	Date:        lipgloss.Color("#81A1C1"), // nord9  - frost, blue
	Hash:        lipgloss.Color("#8FBCBB"), // nord7  - frost, teal
	Ahead:       lipgloss.Color("#A3BE8C"), // nord14 - aurora, green
	Author:      lipgloss.Color("#B48EAD"), // nord15 - frost, purple
}

// themeGruvboxDark maps the slots onto Gruvbox Dark
// (https://github.com/morhetz/gruvbox).
var themeGruvboxDark = Theme{
	Background:  lipgloss.Color("#282828"), // bg0
	FocusBorder: lipgloss.Color("#83A598"), // bright blue
	Cursor:      lipgloss.Color("#458588"), // neutral blue
	Accent:      lipgloss.Color("#FABD2F"), // bright yellow
	Added:       lipgloss.Color("#B8BB26"), // bright green
	Removed:     lipgloss.Color("#FB4934"), // bright red
	Conflict:    lipgloss.Color("#FE8019"), // bright orange
	Error:       lipgloss.Color("#FB4934"), // bright red
	Muted:       lipgloss.Color("#928374"), // gray
	Date:        lipgloss.Color("#83A598"), // bright blue
	Hash:        lipgloss.Color("#8EC07C"), // bright aqua
	Ahead:       lipgloss.Color("#B8BB26"), // bright green
	Author:      lipgloss.Color("#D3869B"), // bright purple
}

// themeDracula maps the slots onto the Dracula palette
// (https://draculatheme.com/).
var themeDracula = Theme{
	Background:  lipgloss.Color("#282A36"), // background
	FocusBorder: lipgloss.Color("#8BE9FD"), // cyan
	Cursor:      lipgloss.Color("#44475A"), // current line
	Accent:      lipgloss.Color("#F1FA8C"), // yellow
	Added:       lipgloss.Color("#50FA7B"), // green
	Removed:     lipgloss.Color("#FF5555"), // red
	Conflict:    lipgloss.Color("#FFB86C"), // orange
	Error:       lipgloss.Color("#FF5555"), // red
	Muted:       lipgloss.Color("#6272A4"), // comment
	Date:        lipgloss.Color("#8BE9FD"), // cyan
	Hash:        lipgloss.Color("#BD93F9"), // purple
	Ahead:       lipgloss.Color("#50FA7B"), // green
	Author:      lipgloss.Color("#FF79C6"), // pink
}

// themeCatppuccinMocha maps the slots onto Catppuccin Mocha
// (https://catppuccin.com/).
var themeCatppuccinMocha = Theme{
	Background:  lipgloss.Color("#1E1E2E"), // base
	FocusBorder: lipgloss.Color("#89B4FA"), // blue
	Cursor:      lipgloss.Color("#45475A"), // surface1
	Accent:      lipgloss.Color("#F9E2AF"), // yellow
	Added:       lipgloss.Color("#A6E3A1"), // green
	Removed:     lipgloss.Color("#F38BA8"), // red
	Conflict:    lipgloss.Color("#FAB387"), // peach
	Error:       lipgloss.Color("#F38BA8"), // red
	Muted:       lipgloss.Color("#6C7086"), // overlay0
	Date:        lipgloss.Color("#89DCEB"), // sky
	Hash:        lipgloss.Color("#94E2D5"), // teal
	Ahead:       lipgloss.Color("#A6E3A1"), // green
	Author:      lipgloss.Color("#CBA6F7"), // mauve
}

// themeTokyoNight maps the slots onto Tokyo Night
// (https://github.com/enkia/tokyo-night-vscode-theme).
var themeTokyoNight = Theme{
	Background:  lipgloss.Color("#1A1B26"), // bg
	FocusBorder: lipgloss.Color("#7AA2F7"), // blue
	Cursor:      lipgloss.Color("#33467C"), // bg highlight
	Accent:      lipgloss.Color("#E0AF68"), // yellow
	Added:       lipgloss.Color("#9ECE6A"), // green
	Removed:     lipgloss.Color("#F7768E"), // red
	Conflict:    lipgloss.Color("#FF9E64"), // orange
	Error:       lipgloss.Color("#F7768E"), // red
	Muted:       lipgloss.Color("#565F89"), // comment
	Date:        lipgloss.Color("#7DCFFF"), // cyan
	Hash:        lipgloss.Color("#73DACA"), // teal
	Ahead:       lipgloss.Color("#9ECE6A"), // green
	Author:      lipgloss.Color("#BB9AF7"), // magenta
}

// themeSolarizedDark maps the slots onto Solarized Dark
// (https://ethanschoonover.com/solarized/).
var themeSolarizedDark = Theme{
	Background:  lipgloss.Color("#002B36"), // base03
	FocusBorder: lipgloss.Color("#268BD2"), // blue
	Cursor:      lipgloss.Color("#073642"), // base02
	Accent:      lipgloss.Color("#B58900"), // yellow
	Added:       lipgloss.Color("#859900"), // green
	Removed:     lipgloss.Color("#DC322F"), // red
	Conflict:    lipgloss.Color("#CB4B16"), // orange
	Error:       lipgloss.Color("#DC322F"), // red
	Muted:       lipgloss.Color("#586E75"), // base01
	Date:        lipgloss.Color("#2AA198"), // cyan
	Hash:        lipgloss.Color("#268BD2"), // blue
	Ahead:       lipgloss.Color("#859900"), // green
	Author:      lipgloss.Color("#D33682"), // magenta
}

// themeOrder is the cycling order for the "t" keybind; themesByName must have
// exactly these keys.
var themeOrder = []string{
	"system",
	"nord",
	"gruvbox",
	"dracula",
	"catppuccin",
	"tokyonight",
	"solarized",
}

var themesByName = map[string]Theme{
	"system":     themeSystem,
	"nord":       themeNord,
	"gruvbox":    themeGruvboxDark,
	"dracula":    themeDracula,
	"catppuccin": themeCatppuccinMocha,
	"tokyonight": themeTokyoNight,
	"solarized":  themeSolarizedDark,
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
