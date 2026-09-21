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

	// Author colors the author initials shown in the log panel.
	Author color.Color
	// Background sets the terminal background color. When nil the
	// terminal's own default background is used.
	Background color.Color
}

// themeSystem relies entirely on the terminal's own ANSI palette (basic
// 4-bit colors plus a couple of 256-color grays), so it inherits whatever
// light/dark scheme the user already has configured instead of imposing
// one.
var themeSystem = Theme{
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
	Author:      lipgloss.ANSIColor(13),
}

// themeNord maps the same slots onto the Nord palette
// (https://www.nordtheme.com/), fixed hex colors instead of terminal ANSI
// slots.
var themeNord = Theme{
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
	Background:  lipgloss.Color("#2E3440"), // nord0  - polar night, darkest
}

// themeDracula maps the slots onto the Dracula palette
// (https://draculatheme.com/).
var themeDracula = Theme{
	FocusBorder: lipgloss.Color("#BD93F9"), // purple
	Cursor:      lipgloss.Color("#44475A"), // current line
	Accent:      lipgloss.Color("#FF79C6"), // pink
	Added:       lipgloss.Color("#50FA7B"), // green
	Removed:     lipgloss.Color("#FF5555"), // red
	Conflict:    lipgloss.Color("#FFB86C"), // orange
	Error:       lipgloss.Color("#FF5555"), // red
	Muted:       lipgloss.Color("#6272A4"), // comment
	Date:        lipgloss.Color("#8BE9FD"), // cyan
	Hash:        lipgloss.Color("#F1FA8C"), // yellow
	Background:  lipgloss.Color("#282A36"), // background
}

// themeCatppuccin maps the slots onto the Catppuccin Mocha palette
// (https://catppuccin.com/).
var themeCatppuccin = Theme{
	FocusBorder: lipgloss.Color("#B4BEFE"), // lavender
	Cursor:      lipgloss.Color("#45475A"), // surface1
	Accent:      lipgloss.Color("#F9E2AF"), // yellow
	Added:       lipgloss.Color("#A6E3A1"), // green
	Removed:     lipgloss.Color("#F38BA8"), // red
	Conflict:    lipgloss.Color("#FAB387"), // peach
	Error:       lipgloss.Color("#F38BA8"), // red
	Muted:       lipgloss.Color("#A6ADC8"), // subtext0
	Date:        lipgloss.Color("#89B4FA"), // blue
	Hash:        lipgloss.Color("#94E2D5"), // teal
	Background:  lipgloss.Color("#1E1E2E"), // base
}

// themeSolarized maps the slots onto the Solarized dark palette
// (https://github.com/altercation/solarized).
var themeSolarized = Theme{
	FocusBorder: lipgloss.Color("#2AA198"), // cyan
	Cursor:      lipgloss.Color("#073642"), // base02
	Accent:      lipgloss.Color("#B58900"), // yellow
	Added:       lipgloss.Color("#859900"), // green
	Removed:     lipgloss.Color("#DC322F"), // red
	Conflict:    lipgloss.Color("#CB4B16"), // orange
	Error:       lipgloss.Color("#DC322F"), // red
	Muted:       lipgloss.Color("#586E75"), // base01
	Date:        lipgloss.Color("#268BD2"), // blue
	Hash:        lipgloss.Color("#839496"), // base0
	Background:  lipgloss.Color("#002B36"), // base03
}

// themeTokyoNight maps the slots onto the Tokyo Night palette
// (https://github.com/enkia/tokyo-night-vscode-theme).
var themeTokyoNight = Theme{
	FocusBorder: lipgloss.Color("#7AA2F7"), // blue
	Cursor:      lipgloss.Color("#414868"), // bg_highlight
	Accent:      lipgloss.Color("#E0AF68"), // yellow
	Added:       lipgloss.Color("#9ECE6A"), // green
	Removed:     lipgloss.Color("#F7768E"), // red
	Conflict:    lipgloss.Color("#FF9E64"), // orange
	Error:       lipgloss.Color("#F7768E"), // red
	Muted:       lipgloss.Color("#565F89"), // comment
	Date:        lipgloss.Color("#7DCFFF"), // cyan
	Hash:        lipgloss.Color("#BB9AF7"), // purple
	Background:  lipgloss.Color("#1A1B26"), // background
}

// themeGruvbox maps the slots onto the Gruvbox dark palette
// (https://github.com/morhetz/gruvbox).
var themeGruvbox = Theme{
	FocusBorder: lipgloss.Color("#83A598"), // blue
	Cursor:      lipgloss.Color("#3C3836"), // bg1
	Accent:      lipgloss.Color("#FABD2F"), // yellow
	Added:       lipgloss.Color("#B8BB26"), // green
	Removed:     lipgloss.Color("#FB4934"), // red
	Conflict:    lipgloss.Color("#FE8019"), // orange
	Error:       lipgloss.Color("#FB4934"), // red
	Muted:       lipgloss.Color("#928374"), // gray
	Date:        lipgloss.Color("#83A598"), // blue
	Hash:        lipgloss.Color("#D3869B"), // purple
	Background:  lipgloss.Color("#282828"), // bg0
}

// themeOneDark maps the slots onto the One Dark palette (Atom editor).
var themeOneDark = Theme{
	FocusBorder: lipgloss.Color("#61AFEF"), // blue
	Cursor:      lipgloss.Color("#3E4452"), // one-dark-bg-highlight
	Accent:      lipgloss.Color("#E5C07B"), // yellow
	Added:       lipgloss.Color("#98C379"), // green
	Removed:     lipgloss.Color("#E06C75"), // red
	Conflict:    lipgloss.Color("#D19A66"), // orange
	Error:       lipgloss.Color("#E06C75"), // red
	Muted:       lipgloss.Color("#5C6370"), // comment
	Date:        lipgloss.Color("#56B6C2"), // cyan
	Hash:        lipgloss.Color("#C678DD"), // purple
	Background:  lipgloss.Color("#282C34"), // background
}

// themeOrder is the cycling order for the "t" keybind; themesByName must have
// exactly these keys.
var themeOrder = []string{
	"system", "nord", "dracula", "catppuccin",
	"solarized", "tokyonight", "gruvbox", "onedark",
}

var themesByName = map[string]Theme{
	"system":     themeSystem,
	"nord":       themeNord,
	"dracula":    themeDracula,
	"catppuccin": themeCatppuccin,
	"solarized":  themeSolarized,
	"tokyonight": themeTokyoNight,
	"gruvbox":    themeGruvbox,
	"onedark":    themeOneDark,
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
