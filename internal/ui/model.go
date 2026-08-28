// Package ui provides the ui for application
package ui

import (
	"time"

	"bubblegit/internal/git"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

const (
	focusStag = iota
	focusBranch
	focusLog
	focusStash
	focusDiff
	focusCount
)

type (
	filesMsg    []git.FileStatus
	branchesMsg []git.BranchInfo
	logMsg      []git.LogEntry
	stashesMsg  []git.StashEntry
	diffMsg     struct{ diff string }
	errMsg      struct{ err error }
	tickMsg     struct{}
)

const (
	commitFocusSummary = iota
	commitFocusMessage
)

type commitPopup struct {
	commitSummary textinput.Model
	commitMessage textarea.Model
	focus         int
	active        bool
	reword        bool
	rewordHash    string
}

type stashBranchPopup struct {
	input  textinput.Model
	active bool
}

// inputAction identifies which git action inputPopup should dispatch on
// submit - one popup, several purposes, same as commitPopup.reword.
type inputAction int

const (
	inputActionNewBranch inputAction = iota
	inputActionRenameBranch
	inputActionPushStash
)

type inputPopup struct {
	input  textinput.Model
	active bool
	action inputAction
	title  string

	// renameFrom is the branch being renamed, captured when the popup is
	// opened. It's read only by inputActionRenameBranch.
	renameFrom string
}

type Model struct {
	dir string

	files       []git.FileStatus
	idxFiles    int
	filesHeight int
	filesWidth  int

	branches     []git.BranchInfo
	idxBranch    int
	branchHeight int
	branchWidth  int

	log       []git.LogEntry
	idxLog    int
	logHeight int
	logWidth  int

	stashes           []git.StashEntry
	idxStash          int
	stashHeight       int
	stashWidth        int
	stashClearConfirm bool

	commitPopup commitPopup

	stashBranchPopup stashBranchPopup

	inputPopup inputPopup

	focus int
	diff  viewport.Model
	err   error

	width  int
	height int

	panelFullScreen bool

	themeName string
	theme     Theme

	ready    bool
	quitting bool
}

func NewModel(dir string) Model {
	return Model{
		dir: dir,
		commitPopup: commitPopup{
			commitSummary: textinput.New(),
			commitMessage: textarea.New(),
		},
		stashBranchPopup: stashBranchPopup{
			input: textinput.New(),
		},
		inputPopup: inputPopup{
			input: textinput.New(),
		},
		themeName: themeOrder[0],
		theme:     themesByName[themeOrder[0]],
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.Refresh(), tickCmd())
}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m Model) Refresh() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			files, err := git.Status(m.dir)
			if err != nil {
				return errMsg{err}
			}
			return filesMsg(files)
		},
		func() tea.Msg {
			branches, err := git.Branches(m.dir)
			if err != nil {
				return errMsg{err}
			}
			return branchesMsg(branches)
		},
		func() tea.Msg {
			log, err := git.Log(m.dir, "HEAD", 200)
			if err != nil {
				return errMsg{err}
			}
			return logMsg(log)
		},
		func() tea.Msg {
			stashes, err := git.Stashes(m.dir)
			if err != nil {
				return errMsg{err}
			}
			return stashesMsg(stashes)
		},
	)
}
